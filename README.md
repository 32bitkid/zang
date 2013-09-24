# Zang

Zang is a simple markdown preprocessor to expand code references from a git respotiory

## Using STDIN and STDOUT

    cat input.md | zang > output.md

This assumes that the current directory is inside of a git repository that contains the referenced commits/files.
If you are rendering from a different location, then you can manually set the location of repository using the `--repo`
flag.

    cat input.md | zang --repo=[path] > output.md

## Markdown syntax

    <!-- {{[format]|git|[refspec]|[path]}} -->

## Basic Example

To reference a file named `foo.bar` at commit `abcd1234`

    <!-- {{csharp|git|abcd1234|foo.bar}} -->

This will get expanded to in the output:

    <!-- {{csharp|git|abcd1234|foo.bar}} -->
    <!-- Begin -->
    ```csharp
    [contents of `foo.bar`]
    ```
    > Commit: abcd1234
    > File: foo.bar
    <!-- End -->

## Using line-numbers

You can also specify a subset of lines to render from the file, by adding `:start:end` at the end of the filename. For
example to render lines 10 to 20 of `foo.bar`:

    <!-- {{csharp|git|abcd1234|foo.bar:10:20}} -->

You can also render a single line from a file by omiting the `:end` portion:

    <!-- {{csharp|git|abcd1234|foo.bar:10}} -->

