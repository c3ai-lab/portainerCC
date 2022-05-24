import angular from 'angular'
import controller from './confidential-computing.controller.js'

angular.module('portainer.app').component('confidentialComputingView', {
  templateUrl: './confidential-computing.html',
  controller,
})
