module.exports = grammar({
  name: 'tau',

  extras: $ => [
    /\s/,
    $.comment,
  ],

  rules: {
    source_file: $ => repeat($._statement),

    _statement: $ => choice(
      $.function_definition,
      $.keyword,
      $.identifier,
      $.literal,
      $.operator,
      $.punctuation
    ),

    comment: $ => token(seq('#', /.*/)),

    function_definition: $ => seq(
      'fn',
      $.identifier,
      $.parameter_list,
      $.block
    ),

    keyword: $ => token(choice(
      'tau', 'if', 'else', 'for', 'return', 'continue', 'break'
    )),

    identifier: $ => token(seq(
      /[a-zA-Z_][a-zA-Z0-9_]*/,
      optional('()')
    )),

    parameter_list: $ => seq(
      '(',
      optional(seq(
        $.identifier,
        repeat(seq(',', $.identifier))
      )),
      ')'
    ),

    block: $ => seq(
      '{',
      repeat($._statement),
      '}'
    ),

    literal: $ => choice(
      $.boolean,
      $.null,
      $.number,
      $.string
    ),

    boolean: $ => token(choice('true', 'false')),

    null: $ => token('null'),

    number: $ => token(choice(
      seq(
        optional('-'),
        choice(
          seq(/[0-9]+/, optional(seq('.', /[0-9]+/))),
          seq('.', /[0-9]+/)
        ),
        optional(seq(
          /[eE]/,
          optional(choice('+', '-')),
          /[0-9]+/
        ))
      ),
      seq('0x', /[0-9a-fA-F]+/),
      seq('0o', /[0-7]+/),
      seq('0b', /[01]+/)
    )),

    string: $ => choice(
      seq(
        '"',
        repeat(choice(
          token.immediate(prec(1, /[^"\\]+/)),
          $.escape_sequence
        )),
        '"'
      ),
      seq(
        '`',
        repeat(choice(
          token.immediate(prec(1, /[^`\\]+/)),
          $.escape_sequence
        )),
        '`'
      )
    ),

    escape_sequence: $ => token(seq(
      '\\',
      choice(
        /[bfnrtv\\"']/,
        /x[0-9a-fA-F]{2}/,
        /u[0-9a-fA-F]{4}/,
        /U[0-9a-fA-F]{8}/,
        /[0-7]{3}/
      )
    )),

    operator: $ => token(choice(
      '<<=', '>>=', '&^=', '&=', '^=', '|=', '%=', '+=', '-=', '*=', '/=', '++', '--',
      '&^', '<<', '>>', '==', '!=', '<=', '>=', '<', '>', '=', '&&', '||', '!', '|', '^', '-', '+', '/', '%'
    )),

    punctuation: $ => token(choice(
      ',', ':', ';', '.', '(', ')', '{', '}', '[', ']'
    )),
  },
});
