var path = require('path');
var merge = require('webpack-merge')
var baseWebpackConfig = require('./webpack.base.config.js')

const isAppProd = process.env.APP_ENV === 'prod'

var webpackConfig = merge(baseWebpackConfig, {
  entry: './src/client.js',
  output: {
    path: path.resolve(__dirname, './public/g'),
    publicPath: isAppProd ? 'https://localhost/static/g/' : '/static/g/',
    filename: 'client.js?[hash:9]',
    chunkFilename: 'chunk.[name].js?[hash:9]'
  },
})
module.exports = webpackConfig
