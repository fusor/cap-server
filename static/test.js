$(function() {
  var socket = io()
  console.log('Trying to connect to server.')

  socket.on('connect', function() {
    console.log('Connected to server');
  })

  socket.on('disconnect', function() {
    console.log('Disconnected from server')
  });

  socket.on('firehose', function(msg) {
    console.log(msg);
  });

  $('button.debug').click(function(e) {
    var uu = url('/nulecules/eriknelson/etherpad-atomicapp/deploy');

    axios.post(uu).then(function(data) {
      console.log('got data');
      console.log(data);
    }).catch(function(err) {
      console.log('failed');
      console.log(err);
    })
  });

});

function url(path) {
  return 'http://cap.example.com:3001/api' + path;
}
