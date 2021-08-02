var path = require('path');
const { merge } = require('webpack-merge');
var baseWebpackConfig = require('./webpack.base.config.js')

const isAppProd = process.env.APP_ENV === 'prod'

var webpackConfig = merge(baseWebpackConfig, {
  target: 'web',
  entry: './src/client.js',
  output: {
    path: path.resolve(__dirname, './dist/g'),
    publicPath: isAppProd ? 'https://localhost/static/g/' : '/static/g/',
    filename: 'client.js?[hash:9]',
    chunkFilename: 'chunk.[name].[hash:9].js'
  },
})
module.exports = webpackConfig
