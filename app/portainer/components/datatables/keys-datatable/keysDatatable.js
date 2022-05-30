angular.module('portainer.app').component('keysDatatable', {
  templateUrl: './keysDatatable.html',
  controller: 'GenericDatatableController',
  bindings: {
    titleText: '@',
    titleIcon: '@',
    dataset: '<',
    tableKey: '@',
    orderBy: '@',
    reverseOrder: '<',
    removeAction: '<',
    teams: '<',
    multiClose: '<',
    multiOpen: '<',
    createAction: '<'
  },
});
