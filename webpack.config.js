const path = require('path');
const webpack = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = {
  entry: {
    vendor: [
      './src/vendor/jquery.min.js',
      './src/vendor/jquery.mousewheel-min.js',
      './src/vendor/keyboard-polyfill-0.1.42.js',
      './src/vendor/jquery.terminal.min.js',
    ],
    app: [
      './src/cmdchallenge.js',
    ],
  },
  output: {
    filename: '[name].js',
    path: path.resolve(__dirname, 'static'),
  },
  performance: {
    hints: false
  },
  module: {
    rules: [
       {
          test: /\.js$/,
          exclude: /(node_modules|bower_components)/,
          use: {
            loader: 'babel-loader',
            options: {
              presets: ['@babel/env']
            }
          }
        },
        {
          test: /jquery.min.js$/,
          loader: 'expose-loader',
          options: {
            exposes: {
              globalName: 'jQuery',
              override: true,
            }
          }
        },
        {
          test: /highlight.+\.js$/,
          loader: 'expose-loader',
          options: {
            exposes: {
              globalName: 'hljs',
              override: true,
            }
          }
        }

    ],
  },
  externals: {
    jquery: 'jQuery',
    hljs: 'hljs',
  },

  plugins: [
    new webpack.ProvidePlugin({
      $: 'jquery',
      jQuery: 'jquery',
      hljs: 'hljs',
    }),
  ]
};
