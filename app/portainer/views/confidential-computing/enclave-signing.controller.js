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
    $state.reload();
    console.log("test");
  }

  this.generateKey = function() {
    KeymanagementService.generateKey("ENCLAVE_SIGNING_KEY", "descriptionxyz").then(function success() {
      console.log("NEW KEY GENERATED");
    }).catch(function error(err) {
      console.log("ERROR");
      console.log(err);
    })
  }

  this.updateKeyAccess = function (key) {
    var newTeamIds = key.teamsSelection.map((team) => { return team.Id })
    if(!_.isEqual(tempTeamIds, newTeamIds)){
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
        var keys = _.orderBy(data.keys, 'name', 'asc');

        $scope.enclaveKeys = keys.map((key) => {
          var savedTeams = [];
          //temp
          if (key.name == "super key") {
            savedTeams.push(data.teams[0]);
          }

          key.teams = angular.copy(data.teams)

          if (savedTeams.length > 0) {
            key.teams = key.teams.map((team) => {
              if (savedTeams.some((saved) => {
                return saved.Id == team.Id;
              })) {
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
