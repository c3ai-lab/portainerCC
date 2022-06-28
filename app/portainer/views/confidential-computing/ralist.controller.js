import angular from 'angular';

angular.module('portainer.app').controller('raController', raController);

/* @ngInject */
export default function raController(Notifications, $q, $scope) {

  $scope.state = {
    actionInProgress: false,
  };

  function initView() {

    var data = [
      {
        timestamp: 1656053594321,
        image: "coolimage:latest",
        mrsigner: "4654AB58FF",
        mrenclave: "33AAFFE111",
      },
      {
        timestamp: 1656053594321,
        image: "coolimage:latest",
        mrsigner: "4654AB58FF",
        mrenclave: "33AAFFE111",
      },
      {
        timestamp: 1656053594321,
        image: "coolimage:latest",
        mrsigner: "4654AB58FF",
        mrenclave: "33AAFFE111",
      },
      {
        timestamp: 1656053594321,
        image: "coolimage:latest",
        mrsigner: "4654AB58FF",
        mrenclave: "33AAFFE111",
      },
      {
        timestamp: 1656053594321,
        image: "coolimage:latest",
        mrsigner: "4654AB58FF",
        mrenclave: "33AAFFE111",
      },
      {
        timestamp: 1656053594321,
        image: "coolimage:latest",
        mrsigner: "4654AB58FF",
        mrenclave: "33AAFFE111",
      },
      {
        timestamp: 1656449204321,
        image: "super:latest",
        mrsigner: "4FFFFFF8FF",
        mrenclave: "33A7777E111",
      }
    ]


    $scope.imageIdentifiers = data;
    console.log("HALLO")
    // $q.all({
    //   keys: KeymanagementService.getKeys(KEY_TYPE),
    //   teams: TeamService.teams()
    // })
    //   .then(function success(data) {
    //     var keys = _.orderBy(data.keys, 'description', 'asc');

    //     $scope.keys = keys.map((key) => {
    //       key.teams = angular.copy(data.teams)

    //       if (!_.isEmpty(key.TeamAccessPolicies)) {
    //         key.teams = key.teams.map((team) => {
    //           if (Object.keys(key.TeamAccessPolicies).includes(team.Id.toString())) {
    //             team.ticked = true;
    //           }
    //           return team;
    //         })
    //       }
    //       return key
    //     })

    //     $scope.teams = _.orderBy(data.teams, 'Name', 'asc');
    //   }).catch(function error(err) {
    //     $scope.keys = [];
    //     $scope.teams = [];
    //     Notifications.error('Failure', err, 'Unable to retrieve keys');
    //   })

  }


  initView();
}
