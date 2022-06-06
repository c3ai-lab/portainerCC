import angular from 'angular';
import _ from 'lodash-es';

angular.module('portainer.app').controller('enclaveSigningController', enclaveSigningController);

/* @ngInject */
export default function enclaveSigningController(Notifications, $q, $scope, KeymanagementService, TeamService, $state, FileSaver) {

  $scope.state = {
    actionInProgress: false,
  };

  const KEY_TYPE = "ENCLAVE_SIGNING_KEY";

  var tempTeamIds = [];

  this.testFunc = function () {
    console.log($scope.formData);
    var teamIds = $scope.formData.teamIds.map((team) => { return team.Id });
    console.log(teamIds);
  }

  $scope.formData = {
    description: "",
    teamIds: []
  }

  this.generateKey = function () {
    $scope.state.actionInProgress = true;

    var teamIds = $scope.formData.teamIds.map((team) => { return team.Id });

    KeymanagementService.generateKey(KEY_TYPE, $scope.formData.description, teamIds)
      .then(function success() {
        Notifications.success('Success', 'New Key added!');
        $state.reload();
      }).catch(function error(err) {
        Notifications.error('Failure', err, 'Unable to generate key');
      })
      .finally(function final() {
        $scope.state.actionInProgress = false;
      });
  }

  this.removeKey = function (selectedKeys) {
    $scope.state.actionInProgress = true;

    $q.all(
      selectedKeys.map(async key => {
        await KeymanagementService.deleteKey(key.id)
          .then(function success() {
            Notifications.success('Success', 'Key deleted!');
          })
          .catch(function error(err) {
            Notifications.error('Failure', err, 'Unable to delete key!');
          })
      })
    )
      .then(function success(data) {
        $scope.state.actionInProgress = false;
        $state.reload();
      })
  }

  this.exportKey = function (selectedKeys) {
    console.log("export keyss")
    console.log(selectedKeys);

    KeymanagementService.getKeyAsPEM(selectedKeys[0].id)
      .then(function success(data) {
        console.log(data);
        var downloadData = new Blob([data.PEM], { type: 'text/plain' });
        FileSaver.saveAs(downloadData, 'enclave_signing_key_' + data.Id + '.pem');
        Notifications.success('Key successfully exported');
      })
      .catch(function error(err) {
        Notifications.error('Failure', err, 'Unable to export key');
      })

  }


  this.importKey = function () {
    console.log("import key")
  }

  this.updateKeyAccess = function (key) {
    var newTeamIds = key.teamsSelection.map((team) => { return team.Id })
    if (!_.isEqual(tempTeamIds, newTeamIds)) {
      $scope.state.actionInProgress = true;
      KeymanagementService.updateTeams(key.id, newTeamIds)
        .then(function success() {
          Notifications.success('Success', 'Access updated!');
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to update access!');
        })
        .finally(function final() {
          $scope.state.actionInProgress = false;
        });
    }
    tempTeamIds = [];
  }

  this.saveTempSelection = function (key) {
    tempTeamIds = key.teamsSelection.map((team) => { return team.Id })
  }

  function initView() {
    $q.all({
      keys: KeymanagementService.getKeys(KEY_TYPE),
      teams: TeamService.teams()
    })
      .then(function success(data) {
        var keys = _.orderBy(data.keys, 'description', 'asc');

        $scope.enclaveKeys = keys.map((key) => {
          key.teams = angular.copy(data.teams)

          if (!_.isEmpty(key.TeamAccessPolicies)) {
            key.teams = key.teams.map((team) => {
              if (Object.keys(key.TeamAccessPolicies).includes(team.Id.toString())) {
                team.ticked = true;
              }
              return team;
            })
          }
          return key
        })

        $scope.teams = _.orderBy(data.teams, 'Name', 'asc');
      }).catch(function error(err) {
        $scope.enclaveKeys = [];
        $scope.teams = [];
        Notifications.error('Failure', err, 'Unable to retrieve keys');
      })

  }


  initView();
}
