var devhubApp = angular.module('devhubApp', []);

devhubApp.controller('ImageListCtrl', function ($scope, $http) {
  $http.get('/api/images').success(function (data) {
    $scope.images = data.images;
  });
  $scope.orderProp = 'age';
});
