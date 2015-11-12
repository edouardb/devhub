var devhubApp = angular.module('devhubApp', ['ngRoute', 'angularMoment']);

devhubApp.controller('MainCtrl', function($scope, $route, $routeParams, $location) {
  $scope.$route = $route;
  $scope.$location = $location;
  $scope.$routeParams = $routeParams;
  $scope.basehref = document.location.protocol + '//' + document.location.host;
});

devhubApp.config(function($routeProvider, $locationProvider) {
  $routeProvider
    .when('/images/:imageId', {
      templateUrl: '/static/image.html',
      controller: 'ImageDetailCtrl'
    })
    .otherwise({
      templateUrl: '/static/home.html',
      controller: 'HomeCtrl'
    });
  // $locationProvider.html5Mode(true);
});

devhubApp.controller('ImageDetailCtrl', function($scope, $http, $routeParams) {
  $scope.name = "ImageDetailCtrl";
  $scope.params = $routeParams;
  $http.get('/api/images/' + $routeParams.imageId).success(function (data) {
    $scope.image = data.image;
  });

  var loadingValue = 'Loading...';
  $scope.dockerfile = loadingValue;
  $scope.makefile = loadingValue;
  $scope.showRawAPI = function() {
    $scope.preRawAPI = true;
    $scope.preDockerfile = false;
    $scope.preMakefile = false;
  };
  $scope.showDockerfile = function() {
    $scope.preRawAPI = false;
    $scope.preDockerfile = true;
    $scope.preMakefile = false;
    if ($scope.dockerfile == loadingValue) {
      $http.get('/api/images/' + $routeParams.imageId + '/dockerfile').success(function (data) {
        $scope.dockerfile = data;
      });
    }
  };
  $scope.showMakefile = function() {
    $scope.preRawAPI = false;
    $scope.preDockerfile = false;
    $scope.preMakefile = true;
    if ($scope.makefile == loadingValue) {
      $http.get('/api/images/' + $routeParams.imageId + '/makefile').success(function (data) {
        $scope.makefile = data;
      });
    }
  };
});

devhubApp.controller('HomeCtrl', function($scope, $http, $routeParams) {
  $scope.name = "HomeCtrl";
  $scope.params = $routeParams;
  $http.get('/api/images').success(function (data) {
    $scope.images = data.images;
  });
  $http.get('/api/bootscripts').success(function (data) {
    $scope.bootscripts = data.bootscripts;
  });
});

devhubApp.filter('prettyJSON', function() {
  function prettyPrintJson(json) {
    if (window.JSON) {
      return JSON.stringify(json, null, '  ');
    }
    return json;
  }
  return prettyPrintJson;
});
