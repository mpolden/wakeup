var wol = wol || {};

wol.state = {
  devices: [],
  deviceToWake: {
    name: '',
    macAddress: ''
  },
  error: {},
  setName: function (v) {
    wol.state.deviceToWake.name = v;
  },
  setMacAddress: function (v) {
    wol.state.deviceToWake.macAddress = v;
  },
  wake: function (device) {
    if (typeof device !== 'undefined') {
      wol.wakeDevice(device);
    } else {
      wol.wakeDevice({name: wol.state.deviceToWake.name,
                      macAddress: wol.state.deviceToWake.macAddress});
    }
  },
  remove: function (device) {
    wol.removeDevice(device);
  },
  add: function (device) {
    var exists = wol.state.devices.some(function (d) {
      return d.macAddress === device.macAddress;
    });
    if (!exists) {
      wol.state.devices.push(device);
      wol.state.devices.sort(function (a, b) {
        if (a.macAddress < b.macAddress) {
          return -1;
        }
        if (a.macAddress > b.macAddress) {
          return 1;
        }
        return 0;
      });
    }
    wol.state.setName('');
    wol.state.setMacAddress('');
    wol.state.error = {};
  }
};

wol.getDevices = function() {
  m.request({method: 'GET', url: '/api/v1/wake'})
    .then(function (data) {
      wol.state.devices = data.devices;
      return data;
    }, function (data) {
      wol.state.error = data;
    });
};

wol.wakeDevice = function(device) {
  m.request({method: 'POST', url: '/api/v1/wake', data: device})
    .then(function (data) {
      wol.state.add(device);
      return data;
    }, function (data) {
      wol.state.error = data;
    });
};

wol.removeDevice = function (device) {
  m.request({method: 'DELETE', url: '/api/v1/wake', data: device})
    .then(function (data) {
      wol.state.devices = wol.state.devices.filter(function (d) {
        return d.macAddress !== device.macAddress;
      });
      return data;
    }, function (data) {
      wol.state.error = data;
    });
};

wol.alertView = function () {
  var e = wol.state.error;
  var isError = Object.keys(e).length !== 0;
  var text = isError ? e.message + ' (' + e.status + ')' : '';
  var cls = 'alert-danger alert-dismissible' + (isError ? '' : ' hidden');
  return m('div.alert', {class: cls, role: 'alert'}, [
    m('span', {class: 'glyphicon glyphicon-exclamation-sign'}),
    m('strong', ' Error: '), text
  ]);
};

wol.devicesView = function () {
  var firstRow = m('tr', [
    m('td', m('input[type=text]', {oninput: m.withAttr('value', wol.state.setName),
                                   value: wol.state.deviceToWake.name,
                                   class: 'form-control',
                                   placeholder: 'Name'})),
    m('td', m('input[type=text]', {oninput: m.withAttr('value', wol.state.setMacAddress),
                                   value: wol.state.deviceToWake.macAddress,
                                   class: 'form-control',
                                   placeholder: 'MAC address'})),
    m('td', m('button[type=button]', {class: 'btn btn-success btn-remove',
                                      onclick: function () { wol.state.wake(); } },
              m('span', {class: 'glyphicon glyphicon-off'}))),
    m('td')
  ]);
  var rows = wol.state.devices.map(function (device) {
    return m('tr', [
      m('td', device.name || ''),
      m('td', m('code', device.macAddress)),
      m('td',
        m('button[type=button]', {class: 'btn btn-success btn-remove',
                                  onclick: function () { wol.state.wake(device); } },
           m('span', {class: 'glyphicon glyphicon-off'})
         )
       ),
      m('td',
        m('button[type=button]', {class: 'btn btn-danger btn-remove',
                                  onclick: function () { wol.state.remove(device); } },
           m('span', {class: 'glyphicon glyphicon-remove'})
         )
       )
    ]);
  });
  return m('table.table', {class: ''},
           m('thead', m('tr', [
             m('th', {class: 'col-md-2'}, 'Device name'),
             m('th', {class: 'col-md-2'}, 'MAC address'),
             m('th', {class: 'col-md-1'}),
             m('th', {class: 'col-md-1'})
           ])),
           m('tbody', [firstRow].concat(rows))
          );
};

wol.oncreate = wol.getDevices();

wol.view = function() {
  return m('div.container', [
    m('div.row', [
      m('div.col-md-6', m('h1', m('span', {class: 'glyphicon glyphicon-flash'}), ' wake-on-lan'))
    ]),
    m('div.row', [
      m('div.col-md-6', wol.alertView())
    ]),
    m('div.row', [
      m('div.col-md-6', wol.devicesView())
    ])
  ]);
};

m.mount(document.getElementById('app'), wol);
