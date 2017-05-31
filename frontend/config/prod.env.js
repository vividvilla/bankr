var merge = require('webpack-merge')
var appEnv = require('./app.env')

merge(appEnv, {
  NODE_ENV: '"production"'
})
