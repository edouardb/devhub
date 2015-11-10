var devhubApp = angular.module('devhubApp', ['ngRoute']);

devhubApp.controller('MainCtrl', function($scope, $route, $routeParams, $location) {
  $scope.$route = $route;
  $scope.$location = $location;
  $scope.$routeParams = $routeParams;
});

devhubApp.config(function($routeProvider, $locationProvider) {
  $routeProvider
    .when('/images/:imageId', {
      templateUrl: '/assets/image.html',
      controller: 'ImageDetailCtrl'
    })
    .otherwise({
      templateUrl: '/assets/home.html',
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
