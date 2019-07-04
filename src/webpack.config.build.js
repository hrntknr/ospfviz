const configBase = require('./webpack.config');

module.exports = Object.assign({}, configBase, {
  mode: 'production',
});
