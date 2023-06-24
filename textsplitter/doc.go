/*
Package textsplitter provides tools for splitting long texts into smaller chunks
based on configurable rules and parameters.
It aims to help in processing these chunks more efficiently
when interacting with language models or other text-processing tools.

The main components of this package are:

- TextSplitter interface: a common interface for splitting texts into smaller chunks.
- RecursiveCharacter: a text splitter that recursively splits texts by different characters (separators)
combined with chunk size and overlap settings.
- Helper functions: utility functions for creating documents out of split texts and rejoining them if necessary.

Using the TextSplitter interface, developers can implement custom
splitting strategies for their specific use cases and requirements.
*/
package textsplitter
