const path = require('path');
const TerserPlugin = require('terser-webpack-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');
const CopyWebpackPlugin = require('copy-webpack-plugin');
const webpack = require('webpack');
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;

// Is production environment
const isProd = process.env.NODE_ENV === 'production';
// Is analyze mode
const isAnalyze = process.argv.includes('--analyze');

module.exports = {
  target: 'node',
  mode: isProd ? 'production' : 'development',
  entry: {
    extension: './src/extension.ts'
  },
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: '[name].js',
    libraryTarget: 'commonjs2',
    clean: true
  },
  devtool: isProd ? 'source-map' : 'eval-source-map',
  externals: {
    vscode: 'commonjs vscode'
  },
  resolve: {
    extensions: ['.ts', '.js']
  },
  module: {
    rules: [
      {
        test: /\.ts$/,
        exclude: /node_modules/,
        use: [
          {
            loader: 'ts-loader',
            options: {
              compilerOptions: {
                module: 'esnext',
                removeComments: isProd
              }
            }
          }
        ]
      }
    ]
  },
  plugins: [
    new CopyWebpackPlugin({
      patterns: [
        {
          from: 'resources',
          to: 'resources'
        }
      ]
    }),
    new webpack.DefinePlugin({
      'process.env.NODE_ENV': JSON.stringify(isProd ? 'production' : 'development'),
      'process.env.IS_PROD': JSON.stringify(isProd)
    })
  ],
  optimization: {
    minimize: isProd,
    minimizer: [
      new TerserPlugin({
        parallel: true,
        terserOptions: {
          ecma: 2020,
          compress: {
            drop_console: false,
            drop_debugger: true,
            passes: 2,
            keep_infinity: true
          },
          format: {
            comments: false,
            ascii_only: true
          }
        }
      })
    ],
    splitChunks: isProd ? {
      chunks: 'all',
      minSize: 30000,
      maxInitialRequests: 3,
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: 'vendor',
          chunks: 'all',
          priority: -10
        },
        default: {
          minChunks: 2,
          priority: -20,
          reuseExistingChunk: true
        }
      }
    } : false
  },
  performance: {
    hints: false
  },
  cache: {
    type: 'memory'
  }
};
