/* eslint-disable prefer-template */
/* eslint-disable no-param-reassign */
// eslint-disable-next-line import/no-extraneous-dependencies
const babel = require("@babel/core");
const path = require("path");
const fs = require("fs");

/**
 *
 * @param {string|Buffer} content Content of the resource file
 * @param {object} [map] SourceMap data consumable by https://github.com/mozilla/source-map
 * @param {any} [meta] Meta data, could be anything
 */
async function webpackLoader(content, map, meta) {
  const cb = this.async();

  try {
    cb(null, JSON.stringify({ content, imports: [] }), map, meta);
  } catch (err) {
    cb(err)
  }

  /* if (!this.resourcePath.endsWith(".go")) {
    cb(null, JSON.stringify({ content, imports: [] }), map, meta);
    return;
  } */

  /* console.log(this.resourcePath)

  try {
    cb(null, JSON.stringify({ contentcontent, map, meta);
  } catch (err) {
    cb(err);
  } */
}

module.exports = webpackLoader;
