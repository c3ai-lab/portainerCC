import { ImageBuildModel } from '../models/image';

angular.module('portainer.docker').factory('BuildService', [
  '$q',
  'Build',
  'FileUploadService',
  function BuildServiceFactory($q, Build, FileUploadService) {
    'use strict';
    var service = {};

    service.buildImageFromUpload = function (names, file, path, signingKey, inputDir, modelDir) {
      var deferred = $q.defer();

      FileUploadService.buildImage(names, file, path, signingKey, inputDir, modelDir)
        .then(function success(response) {
          var model = new ImageBuildModel(response.data);
          deferred.resolve(model);
        })
        .catch(function error(err) {
          deferred.reject(err);
        });

      return deferred.promise;
    };

    service.buildImageFromURL = function (names, url, path, signingKey, inputDir, modelDir) {
      var params = {
        t: names,
        remote: url,
        dockerfile: path,
        signingKeyId: signingKey,
        inputVolume: inputDir,
        modelVolume: modelDir
      };

      var deferred = $q.defer();

      Build.buildImage(params, {})
        .$promise.then(function success(data) {
          var model = new ImageBuildModel(data);
          deferred.resolve(model);
        })
        .catch(function error(err) {
          deferred.reject(err);
        });

      return deferred.promise;
    };

    service.buildImageFromDockerfileContent = function (names, content, signingKey, inputDir, modelDir) {
      var params = {
        t: names,
      };
      var payload = {
        content: content,
        signingKeyId: signingKey,
        inputVolume: inputDir,
        modelVolume: modelDir
      };

      var deferred = $q.defer();

      Build.buildImageOverride(params, payload)
        .$promise.then(function success(data) {
          var model = new ImageBuildModel(data);
          deferred.resolve(model);
        })
        .catch(function error(err) {
          deferred.reject(err);
        });

      return deferred.promise;
    };

    return service;
  },
]);
