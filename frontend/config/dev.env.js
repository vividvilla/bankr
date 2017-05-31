var merge = require('webpack-merge')
var prodEnv = require('./prod.env')
var appEnv = require('./app.env')

module.exports = merge(prodEnv, appEnv, {
  NODE_ENV: '"development"'
})
