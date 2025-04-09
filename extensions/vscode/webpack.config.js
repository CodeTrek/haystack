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
    extension: './src/extension.ts',
    // No longer use search.js as an entry point, we use the original file directly
    // 'resources/search': './resources/search.js'
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
    extensions: ['.ts', '.js'],
    // Prefer smaller modules
    mainFields: ['module', 'main']
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
                module: 'esnext', // Use esnext to enable better tree-shaking
                removeComments: isProd // Remove comments in production
              },
              transpileOnly: true
            }
          }
        ]
      },
      {
        test: /\.css$/,
        use: ['style-loader', 'css-loader']
      }
    ]
  },
  plugins: [
    // Copy static resources - include JS files as well
    new CopyWebpackPlugin({
      patterns: [
        {
          from: 'resources',
          to: 'resources'
          // No ignore patterns - include all files
        }
      ]
    }),
    // Define environment variables for optimization
    new webpack.DefinePlugin({
      'process.env.NODE_ENV': JSON.stringify(isProd ? 'production' : 'development'),
      'process.env.IS_PROD': JSON.stringify(isProd)
    }),
    // Ignore moment.js locale directory in production
    isProd ? new webpack.IgnorePlugin({
      resourceRegExp: /^\.\/locale$/,
      contextRegExp: /moment$/
    }) : null,
    // Only add analyzer plugin in analyze mode
    isAnalyze ? new BundleAnalyzerPlugin() : null
  ].filter(Boolean),
  optimization: {
    // Enable tree-shaking
    usedExports: true,
    // Optimize module splitting in production
    splitChunks: isProd ? {
      chunks: 'all',
      maxInitialRequests: Infinity,
      minSize: 0,
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name(module) {
            // Split node_modules modules into different bundles based on path
            const packageName = module.context.match(/[\\/]node_modules[\\/](.*?)([\\/]|$)/)[1];
            return `vendor.${packageName.replace('@', '')}`;
          }
        }
      }
    } : false,
    minimize: isProd,
    minimizer: [
      new TerserPlugin({
        parallel: true, // Parallel compression
        terserOptions: {
          ecma: 2020,
          parse: {},
          compress: {
            drop_console: false, // Keep console.log statements
            drop_debugger: true,
            pure_funcs: [], // Don't remove any functions
            passes: 2, // Multiple compression passes
            booleans_as_integers: true,
            keep_infinity: true
          },
          mangle: {
            properties: false
          },
          format: {
            comments: false,
            ascii_only: true
          }
        }
      }),
      new CssMinimizerPlugin()
    ]
  },
  performance: {
    hints: false
  },
  cache: {
    type: 'memory'
  }
};
