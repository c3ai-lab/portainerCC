import angular from 'angular';

angular.module('portainer.app').controller('enclaveSigningController', enclaveSigningController);

/* @ngInject */
export default function enclaveSigningController(Notifications, $async, $http, $q, $scope) {
  var ctrl = this;

  var deferred = $q.defer();
  console.log(ctrl);
  console.log(deferred);

  // TODO in service auslagern

  this.generateKey = () => {
    console.log('--------------------genKEy');
    $async(async () => {
      $http
        .post('https://localhost:9443/api/settings/sgx-keygen', { name: 'superGeilerSigninKey' })
        .then(function success(data) {
          Notifications.success('Success', 'New SGX Signing Key created!');
          console.log(data);
        })
        .catch(function error(err) {
          console.log(err);
        });
      console.log('async');
    });
  };

  this.importKey = () => {
    Notifications.error('Failure', 'No import implemented yet!');
    console.log('-------------------------importkey');
    $async(async () => {
      $http
        .get('https://localhost:9443/api/users')
        .then(function success(data) {
          console.log('ja moin');
          console.log(data);
        })
        .catch(function error(err) {
          console.log(err);
        });
      console.log('async');
    });
  };

  function initView(){
    console.log("MOIN");
    $scope.keys = [{name:"name1"}]
  }

  initView();
}
