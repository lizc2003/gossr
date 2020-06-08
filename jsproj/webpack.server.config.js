var path = require('path')
var merge = require('webpack-merge')
var baseWebpackConfig = require('./webpack.base.config.js')

var webpackConfig = merge(baseWebpackConfig, {
  target: 'node',
  entry: {
    app: './src/server.js'
  },
  output: {
    path: path.resolve(__dirname, './server_dist/g'),
    filename: 'server.js',
    chunkFilename: 'chunk.[name].js',
    libraryTarget: 'commonjs2'
  },
  devtool: 'cheap-source-map',
  module: {
    rules: [
      {
        test: /\.css$/,
        use: [
          'css-loader',
        ]
      },
    ]
  },
})
module.exports = webpackConfig
