const path = require('path');
const VueLoaderPlugin = require('vue-loader/lib/plugin');
const TerserPlugin = require('terser-webpack-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin')
const ProgressBarPlugin  = require('progress-bar-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const autoprefixer = require('autoprefixer');
const pxtorem = require('postcss-pxtorem');

const isProd = process.env.NODE_ENV === 'production'

module.exports = {
  mode: isProd ? 'production' : 'development',
  module: {
    rules: [
      {
        test: /\.vue$/,
        use:[
          {
            loader: 'vue-loader',
            options: {
              // enable CSS extraction
              // extractCSS: true
            }
          }
        ],
      },
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use:[
          {loader: 'babel-loader'}
        ]
      },
      {
        test: /\.css$/,
        use: [
          MiniCssExtractPlugin.loader,
          'css-loader',
        ]
      },
      {
        test: /\.scss$/,
        use:[
          'vue-style-loader',
          'css-loader',
          {
            loader: 'postcss-loader',
            options: {
              postcssOptions:{
                plugins: [
                  autoprefixer({}),
                  pxtorem({
                    rootValue: 75,
                    unitPrecision: 8,
                    propList: ['*'],
                    selectorBlackList: ['.norem'],
                    replace: true,
                    mediaQuery: false,
                    minPixelValue: 1
                  }),
                ],
              }
            }
          },
          {
            loader: 'sass-loader',
            options: {
            },
          },
        ]
      },
      {
        test: /\.(png|jpg|gif|svg|ttf|woff2|woff|eot)$/,
        loader: 'url-loader',
        options: {
          limit: 10000, // encode base64 within 10K
          name: 'img/[name].[hash:9].[ext]'
        }
      },
    ],
  },
  plugins: [
    new VueLoaderPlugin(),
    new MiniCssExtractPlugin({
      filename: 'css/main.css',
      chunkFilename: 'css/chunk.[id].css',
      ignoreOrder: false,
    }),
    new ProgressBarPlugin({
      format: 'build [:bar] :percent (:elapsed seconds)',
      clear: false,
      width: 60
    }),
  ],
  resolve: {
    alias: {
      'vue$': 'vue/dist/vue.common.js',
      '@': path.resolve('src')
    },
    modules: [ path.resolve(__dirname,'node_modules') ],
    extensions: ['*', '.js', '.vue', '.json','.scss']
  },
  devServer: {
    historyApiFallback: true,
    noInfo: true,
    overlay: true,
    hot: true
  },
  performance: {
    hints: false
  },
  devtool: 'eval-source-map',
}

if (isProd) {
  module.exports.optimization = {
    minimize: true,
    minimizer: [new TerserPlugin(), new CssMinimizerPlugin()],
  }

  module.exports.devtool = 'cheap-source-map'
}
