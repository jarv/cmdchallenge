/* eslint-disable global-require */

'use strict';

const shippedProposalsPlugins = () => [
  require('@babel/plugin-syntax-numeric-separator').default,
];

const checkBooleanOptions = (opts) => {
  ['loose', 'shippedProposals'].forEach((optionName) => {
    if (
      opts[optionName] !== undefined &&
      typeof opts[optionName] !== 'boolean'
    ) {
      throw new Error(
        `Preset modern-browsers '${optionName}' option must be a boolean.`,
      );
    }
  });
};

const checkRemovedOptions = (opts) => {
  [
    'esnext',
    'safari10',
    'edge',
    'es2016',
    'es2017',
    'es2018',
    'es2019',
    'supportVariablesFunctionName',
  ].forEach((optionName) => {
    if (opts[optionName] !== undefined) {
      throw new Error(
        `Preset modern-browsers '${optionName}' option was removed`,
      );
    }
  });
};

module.exports = function preset(context, opts = {}) {
  checkBooleanOptions(opts);
  checkRemovedOptions(opts);

  const modules = opts.modules !== undefined ? opts.modules : 'commonjs';

  if (modules !== false && modules !== 'commonjs') {
    throw new Error(
      "Preset modern-browsers 'modules' option must be 'false' to indicate no modules\n" +
        "or 'commonjs' (default)",
    );
  }

  const loose = opts.loose !== undefined ? opts.loose : false;
  const shippedProposals =
    opts.shippedProposals !== undefined ? opts.shippedProposals : true;

  const optsLoose = { loose };

  return {
    plugins: [
      /* es2015 */
      modules === 'commonjs' && [
        require('@babel/plugin-transform-modules-commonjs'),
        optsLoose,
      ],

      /* shippedProposals */
      ...(shippedProposals ? shippedProposalsPlugins() : []),
    ].filter(Boolean),
  };
};
