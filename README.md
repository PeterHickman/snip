# snip - access code snippets from the cli

When I hit a problem I Google for a solution and after finding what I need (probably on Stackoverflow) I end up with a small snippet of code that I squirrel away in a file somewhere on my computer

The keyword here is _somewhere_! I then have to look all over several drives to find it. I know I kept it, but the where eludes me. So I will probably end up on Stackoverflow again, create another small text file and puts it somewhere to keep it safe

Rinse and repeat

Finally I decided to write a tool to keep this in order and stuff it in a sqlite database that can be updated and searched. So I gathered up all the snippets I could find and wrote some code. And here it is, `snip`

```bash
$ snip --import text_summarization_in_python.txt
Imported [text summarization in python]
Imported 1 files, skipped 0 existing
```

Will import the contents of the file `text_summarization_in_python.txt` using the filename as the title. You can import as many files are you want in a single command. It will skip importing files that have the same title / filename. Lets list all the snippets in the database

```bash
$ snip --list
#1 : text summarization in python
```
And display the content with

```bash
$ snip 1
#1 : text summarization in python

# coding=UTF-8
from __future__ import division
import re

# This is a naive text summarization algorithm
# Created by Shlomi Babluki
# April, 2013

class SummaryTool(object):
    # Naive method for splitting a text into sentences
    def split_content_to_sentences(self, content):
        content = content.replace("\n", ". ")
        return content.split(". ")
...
```
When the number of snippets gets large enough there is a basic search option

```bash
$ snip --search text python
#1 (1.000000) : text summarization in python
```

If you need to remove a snippet

```bash
$ snip --delete 1
```

Will do the job and you can export the whole database into a directory as individual files

```bash
$ snip --export new_dir/
```

Still haven't found that bash script I squirrelled away to set ansi colours. I'm sure I'll find it again eventually (or I'll Google it again)

