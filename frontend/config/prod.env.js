var merge = require('webpack-merge')
var config = require('./config.env')

module.exports =  merge(config, {
  NODE_ENV: '"production"'
})
