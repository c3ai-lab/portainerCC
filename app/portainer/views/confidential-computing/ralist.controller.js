import angular from 'angular';
import _ from 'lodash-es';

angular.module('portainer.app').controller('raController', raController);

/* @ngInject */
export default function raController(Notifications, $q, $scope, RaService) {

  $scope.state = {
    actionInProgress: false,
  };

  function initView() {

    $q.all({
      images: RaService.getImages(),
    })
      .then(function success(data) {
        $scope.imageIdentifiers = _.orderBy(data.images, 'image', 'asc');
        console.log("MOIN");
        console.log($scope.data)
      }).catch(function error(err) {
        $scope.imageIdentifiers = [];
        Notifications.error('Failure', err, 'Unable to retrieve keys');
      })

  }


  initView();
}
