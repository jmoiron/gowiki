/* add prettyprint + linenums to codeblocks */
$(function() {
    // old-style pre[lang] prettyprinter from outdated blackfriday
    $("pre[lang]").each(function() {
        $(this).addClass("prettyprint");
    });
    // converts the output from shurcooL ghm to code blocks
    $("div.highlight").replaceWith(function() {
        var classes = $(this).attr("class");
        return $(this)
            .contents()
            .attr({class: classes})
            .addClass("prettyprint");
    });
});

/* insert spaces instead of tabs in textareas */
$(document).delegate('textarea', 'keydown', function(e) {
    var keyCode = e.keyCode || e.which;

    if (keyCode == 9) {
        e.preventDefault();
        var start = $(this).get(0).selectionStart;
        var end = $(this).get(0).selectionEnd;

        // set textarea value to: text before caret + tab + text after caret
        $(this).val($(this).val().substring(0, start)
                    + "    "
                    + $(this).val().substring(end));

        // put caret at right position again
        $(this).get(0).selectionStart =
        $(this).get(0).selectionEnd = start + 4;
    }
});
/* expand/collapse page info */
$(function() {
    $("#page-info-toggle").on("click", function(e) {
        e.preventDefault();
        $("#page-info").toggle();
    });
});
