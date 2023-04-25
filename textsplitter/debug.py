def _join_docs(docs, separator: str):
        text = separator.join(docs)
        text = text.strip()
        if text == "":
            return None
        else:
            return text


def _merge_splits(splits, separator: str, chunk_size, overlap):
        print("merging", splits)
        print("sep len", len(separator))
        # We now want to combine these smaller pieces into medium size
        # chunks to send to the LLM.
        separator_len = len(separator)

        docs = []
        current_doc = []
        total = 0
        for d in splits:
            print("current doc start ", current_doc)
            _len = len(d)
            print("total + _len + (separator_len if len(current_doc) > 0 else 0)>chunk_size:", total + _len + (separator_len if len(current_doc) > 0 else 0)>chunk_size, total, _len, (separator_len if len(current_doc) > 0 else 0), chunk_size)
            if (
                total + _len + (separator_len if len(current_doc) > 0 else 0)
                >chunk_size
            ):
                if total > chunk_size:
                    pass
                if len(current_doc) > 0:
                    doc = _join_docs(current_doc, separator)
                    if doc is not None:
                        docs.append(doc)
                    # Keep on popping if:
                    # - we have a larger chunk than in the chunk overlap
                    # - or if we still have any chunks and the length is long
                    while total > overlap or (
                        total + _len + (separator_len if len(current_doc) > 0 else 0)
                        > chunk_size
                        and total > 0
                    ):
                        print("current doc before pop", current_doc)
                        total -= len(current_doc[0]) + (
                            separator_len if len(current_doc) > 1 else 0
                        )
                        current_doc = current_doc[1:]
                        print("current doc after pop", current_doc)
            current_doc.append(d)
            print("adding to total ",  _len + (separator_len if len(current_doc) > 1 else 0), "len", _len, "sep", (separator_len if len(current_doc) > 1 else 0))
            total += _len + (separator_len if len(current_doc) > 1 else 0)
        doc = _join_docs(current_doc, separator)
        if doc is not None:
            docs.append(doc)
        print("result: ", docs)
        return docs

def split_text(text: str, separators, chunk_size, chunkOverlap):
        """Split incoming text and return chunks."""
        final_chunks = []
        # Get appropriate separator to use
        separator = separators[-1]
        for _s in separators:
            if _s == "":
                separator = _s
                break
            if _s in text:
                separator = _s
                break
        # Now that we have the separator, split the text
        if separator:
            splits = text.split(separator)
        else:
            splits = list(text)
        # Now go merging things, recursively splitting longer texts.
        _good_splits = []
        for s in splits:
            if len(s) < chunk_size:
                _good_splits.append(s)
            else:
                if _good_splits:
                    merged_text = _merge_splits(_good_splits, separator, chunk_size, chunkOverlap)
                    final_chunks.extend(merged_text)
                    _good_splits = []
                other_info = split_text(s, separators, chunk_size, chunkOverlap)
                final_chunks.extend(other_info)
        if _good_splits:
            merged_text =_merge_splits(_good_splits, separator, chunk_size, chunkOverlap)
            final_chunks.extend(merged_text)
        return final_chunks
        
split = split_text("""Hi.\n\nI'm Harrison.\n\nHow? Are? You?\nOkay then f f f f.
This is a weird text to write, but gotta test the splittingggg some how.

Bye!\n\n-H.""", ["\n\n", "\n", " ", ""], 10, 1)

print(split)