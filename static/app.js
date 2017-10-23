var wol = wol || {};

wol.state = {
  devices: [],
  toWake: {
    name: '',
    macAddress: '',
    setName: function(v) {
      wol.state.toWake.name = v;
    },
    setMacAddress: function(v) {
      wol.state.toWake.macAddress = v;
    },
  },
  success: {
    timeout: null,
    device: {}
  },
  error: {},
  wake: function (device) {
    if (typeof device !== 'undefined') {
      wol.wakeDevice(device);
    } else {
      // Copy the toWake object here to avoid input values binding
      wol.wakeDevice({name: wol.state.toWake.name,
                      macAddress: wol.state.toWake.macAddress});
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
    wol.state.toWake.setName('');
    wol.state.toWake.setMacAddress('');
    wol.state.error = {};
  },
  setSuccess: function (device) {
    wol.state.success.device = device;
    // Clear any pending timeout so that timeout is extended for each new
    // wake-up
    clearTimeout(wol.state.success.timeout);
    wol.state.success.timeout = setTimeout(function () {
      wol.state.success.device = {};
      m.redraw();
    }, 4000);
  }
};

wol.getDevices = function() {
  m.request({method: 'GET', url: 'api/v1/wake'})
    .then(function (data) {
      wol.state.devices = data.devices;
      return data;
    }, function (data) {
      wol.state.error = data;
    });
};

wol.wakeDevice = function(device) {
  m.request({method: 'POST', url: 'api/v1/wake', data: device})
    .then(function (data) {
      wol.state.add(device);
      wol.state.setSuccess(device);
      return data;
    }, function (data) {
      wol.state.error = data;
    });
};

wol.removeDevice = function (device) {
  m.request({method: 'DELETE', url: 'api/v1/wake', data: device})
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
  var cls = 'alert-danger' + (isError ? '' : ' hidden');
  return m('div.alert', {class: cls}, [
    m('span', {class: 'glyphicon glyphicon-exclamation-sign'}),
    m('strong', ' Error: '), text
  ]);
};

wol.successView = function () {
  var device = wol.state.success.device;
  var isSuccess = Object.keys(device).length !== 0;
  var name = device.name ? ' (' + device.name + ')' : '';
  var cls = 'alert-success' + (isSuccess ? '' : ' hidden');
  return m('div.alert', {class: cls}, [
    m('span', {class: 'glyphicon glyphicon-ok'}),
    ' Successfully woke ', m('strong', device.macAddress), name
  ]);
};

wol.devicesView = function () {
  var form = m('form', {
    id: 'wake-form',
    onsubmit: function (e) {
      e.preventDefault();
      wol.state.wake();
    }
  });
  var firstRow = m('tr', [
    m('td', m('input[type=text]', {'form': form.attrs.id,
                                   onchange: m.withAttr('value', wol.state.toWake.setName),
                                   value: wol.state.toWake.name,
                                   class: 'form-control',
                                   placeholder: 'Name'})),
    m('td', m('input[type=text]', {'form': form.attrs.id,
                                   onchange: m.withAttr('value', wol.state.toWake.setMacAddress),
                                   value: wol.state.toWake.macAddress,
                                   class: 'form-control',
                                   placeholder: 'MAC address'})),
    m('td', m('button[type=submit]',
              {'form': form.attrs.id, class: 'btn btn-success btn-remove'},
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
  return [form,
          m('table.table', {class: ''},
            m('thead', m('tr', [
              m('th', {class: 'col-md-2'}, 'Device name'),
              m('th', {class: 'col-md-2'}, 'MAC address'),
              m('th', {class: 'col-md-1'}),
              m('th', {class: 'col-md-1'})
            ])),
            m('tbody', [firstRow].concat(rows))
           )];
};

wol.oncreate = wol.getDevices;

wol.view = function() {
  return m('div.container', [
    m('div.row',
      m('div.col-md-6', m('h1', m('span', {class: 'glyphicon glyphicon-flash'}), ' wake-on-lan'))
    ),
    m('div.row', m('div.col-md-6', wol.alertView())),
    m('div.row', m('div.col-md-6', wol.successView())),
    m('div.row', m('div.col-md-6', wol.devicesView()))
  ]);
};

m.mount(document.getElementById('app'), wol);
