import angular from 'angular';

angular.module('portainer.app').controller('ConfidentialComputingController', ConfidentialComputingController);

/* @ngInject */
export default function ConfidentialComputingController(Notifications, $async, $http, $q) {

  var ctrl = this;

  var deferred = $q.defer();

  // TODO in service auslagern

  this.generateKey = () => {
    Notifications.success('Success', 'New SGX Signing Key created!');
    console.log("--------------------genKEy");
    $async(async () => {
      $http.post('https://localhost:9443/api/settings/sgx-keygen', { name: 'superGeilerSigninKey' }).then(function success(data) {
        console.log("ja moin");
        console.log(data);
      }).catch(function error(err) {
        console.log("ERRORRRRRRRRRRRRRRR");
      })
      console.log("async");
    })
  }

  this.importKey = () => {
    Notifications.error('Failure', 'No import implemented yet!');
    console.log("-------------------------importkey");
    $async(async () => {
      $http.get('https://localhost:9443/api/users').then(function success(data) {
        console.log("ja moin");
        console.log(data);
      }).catch(function error(err) {
        console.log("ERRORRRRRRRRRRRRRRR");
      })
      console.log("async");
    })
  }
}