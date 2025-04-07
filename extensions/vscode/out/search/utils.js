"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.parseSearchQuery = parseSearchQuery;
exports.formatSearchResults = formatSearchResults;
exports.getSearchOptions = getSearchOptions;
function parseSearchQuery(query) {
    return query.trim().toLowerCase();
}
function formatSearchResults(results) {
    return results;
}
function getSearchOptions() {
    return {
        includePattern: { '**/*': true },
        excludePattern: { '**/node_modules/**': true }
    };
}
