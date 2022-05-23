import angular from 'angular';

angular.module('portainer.app').controller('ConfidentialComputingController', ConfidentialComputingController);

/* @ngInject */
export default function ConfidentialComputingController(Notifications){

  var ctrl = this;

  this.generateKey = () => {
    Notifications.success('Success', 'New SGX Signing Key created!');
    console.log("--------------------genKEy");
  }

  this.importKey = () => {
    Notifications.success('Success', 'New SGX Signing Key imported!');
    console.log("-------------------------importkey");
  }
}