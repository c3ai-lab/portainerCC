import angular from 'angular';
import _ from 'lodash-es';

angular.module('portainer.app').controller('enclaveSigningController', enclaveSigningController);

/* @ngInject */
export default function enclaveSigningController(Notifications, $async, $http, $q, $scope, KeymanagementService, TeamService, $state) {
  var ctrl = this;
  var deferred = $q.defer();

  var tempTeamIds = [];

  console.log(ctrl);
  console.log(deferred);

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
    var teamIds = $scope.formData.teamIds.map((team) => { return team.Id });
    $q.all(
      KeymanagementService.generateKey("ENCLAVE_SIGNING_KEY", $scope.formData.description, teamIds)
    )
      .then(function success() {
        Notifications.success('Success', 'New Key added!')
      }).catch(function error(err) {
        Notifications.error('Failure', err, 'Unable to generate key');
      })
  }

  this.updateKeyAccess = function (key) {
    var newTeamIds = key.teamsSelection.map((team) => { return team.Id })
    if (!_.isEqual(tempTeamIds, newTeamIds)) {
      Notifications.success('Success', 'Access for Key ' + key.id + ' updated!');
      console.log("UPDATE KEY!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!");
    }
    tempTeamIds = [];
  }

  this.saveTempSelection = function (key) {
    tempTeamIds = key.teamsSelection.map((team) => { return team.Id })
  }

  function initView() {
    $q.all({
      keys: KeymanagementService.getKeys("coolType"),
      teams: TeamService.teams()
    })
      .then(function success(data) {
        var keys = _.orderBy(data.keys, 'description', 'asc');

        $scope.enclaveKeys = keys.map((key) => {
          // //temp
          // if (key.name == "super key") {
          //   savedTeams.push(data.teams[0]);
          // }

          key.teams = angular.copy(data.teams)

          if (key.teamIds?.length > 0) {
            key.teams = key.teams.map((team) => {
              if (key.teamIds.includes(team.Id)) {
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
