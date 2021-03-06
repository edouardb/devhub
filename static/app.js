var devhubApp = angular.module('devhubApp', ['ngRoute', 'angularMoment', 'ngSanitize', 'angular-loading-bar', 'ngAnimate']);

devhubApp.controller('MainCtrl', function($scope, $route, $routeParams, $location) {
  $scope.$route = $route;
  $scope.$location = $location;
  $scope.$routeParams = $routeParams;
  $scope.basehref = document.location.protocol + '//' + document.location.host;
});

devhubApp.config(function($routeProvider, $locationProvider) {
  $routeProvider
    .when('/images', {
      templateUrl: '/static/images.html',
      controller: 'ImageListCtrl'
    })
    .when('/bootscripts', {
      templateUrl: '/static/bootscripts.html',
      controller: 'BootscriptListCtrl'
    })
    .when('/bootscripts/:bootscriptId', {
      templateUrl: '/static/bootscript.html',
      controller: 'BootscriptDetailCtrl'
    })
    .when('/images/:imageId', {
      templateUrl: '/static/image.html',
      controller: 'ImageDetailCtrl'
    })
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

devhubApp.controller('BootscriptDetailCtrl', function($scope, $http, $routeParams) {
  $scope.name = "ImageDetailCtrl";
  $scope.params = $routeParams;
  $http.get('/api/bootscripts/' + $routeParams.bootscriptId).success(function (data) {
    $scope.bootscript = data.bootscript;
  });
  $scope.showRawAPI = function() {
    $scope.preRawAPI = true;
  };
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

devhubApp.controller('ImageListCtrl', function($scope, $http, $routeParams) {
  $scope.orderByField = 'objects.api.creation_date';
  $scope.reverseSort = true;
  $http.get('/api/images').success(function (data) {
    $scope.images = data.images;
  });
});

devhubApp.controller('BootscriptListCtrl', function($scope, $http, $routeParams) {
  $scope.orderByField = 'title';
  $scope.reverseSort = true;
  $http.get('/api/bootscripts').success(function (data) {
    $scope.bootscripts = data.bootscripts;
  });
});

devhubApp.controller('HomeCtrl', function($scope, $http, $routeParams) {
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

devhubApp.filter('addLabels', function() {
  return function(text) {

    text = text.replace(/\(?beta\)?/i, '<span class="inline-block px1 white bg-black rounded">beta</span>');
    text = text.replace(/\(?stable\)?/i, '<span class="inline-block px1 white bg-green rounded">stable</span>');
    text = text.replace(/\(?lts\)?/i, '<span class="inline-block px1 white bg-blue rounded">lts</span>');
    text = text.replace(/\(?latest\)?/i, '<span class="inline-block px1 white bg-purple rounded">latest</span>');

    return text;
  };
});

devhubApp.filter('humanBytes', function() {
  return function(bytes, precision, base) {
    if (isNaN(parseFloat(bytes)) || !isFinite(bytes)) return '-';
    if (typeof precision === 'undefined') precision = 1;
    if (typeof base === 'undefined') base = 2;
    var units = ['bytes', 'kB', 'MB', 'GB', 'TB', 'PB'],
        number = Math.floor(Math.log(bytes) / Math.log(1024));
    var ceil = base > 2 ? 1000 : 1024;
    return (bytes / Math.pow(ceil, Math.floor(number))).toFixed(precision) +  ' ' + units[number];
  };
});
