#include "tree_sitter/parser.h"

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 14
#define STATE_COUNT 47
#define LARGE_STATE_COUNT 9
#define SYMBOL_COUNT 31
#define ALIAS_COUNT 0
#define TOKEN_COUNT 20
#define EXTERNAL_TOKEN_COUNT 0
#define FIELD_COUNT 0
#define MAX_ALIAS_SEQUENCE_LENGTH 4
#define PRODUCTION_ID_COUNT 1

enum ts_symbol_identifiers {
  sym_comment = 1,
  anon_sym_fn = 2,
  sym_keyword = 3,
  sym_identifier = 4,
  anon_sym_LPAREN = 5,
  anon_sym_COMMA = 6,
  anon_sym_RPAREN = 7,
  anon_sym_LBRACE = 8,
  anon_sym_RBRACE = 9,
  sym_boolean = 10,
  sym_null = 11,
  sym_number = 12,
  anon_sym_DQUOTE = 13,
  aux_sym_string_token1 = 14,
  anon_sym_BQUOTE = 15,
  aux_sym_string_token2 = 16,
  sym_escape_sequence = 17,
  sym_operator = 18,
  sym_punctuation = 19,
  sym_source_file = 20,
  sym__statement = 21,
  sym_function_definition = 22,
  sym_parameter_list = 23,
  sym_block = 24,
  sym_literal = 25,
  sym_string = 26,
  aux_sym_source_file_repeat1 = 27,
  aux_sym_parameter_list_repeat1 = 28,
  aux_sym_string_repeat1 = 29,
  aux_sym_string_repeat2 = 30,
};

static const char * const ts_symbol_names[] = {
  [ts_builtin_sym_end] = "end",
  [sym_comment] = "comment",
  [anon_sym_fn] = "fn",
  [sym_keyword] = "keyword",
  [sym_identifier] = "identifier",
  [anon_sym_LPAREN] = "(",
  [anon_sym_COMMA] = ",",
  [anon_sym_RPAREN] = ")",
  [anon_sym_LBRACE] = "{",
  [anon_sym_RBRACE] = "}",
  [sym_boolean] = "boolean",
  [sym_null] = "null",
  [sym_number] = "number",
  [anon_sym_DQUOTE] = "\"",
  [aux_sym_string_token1] = "string_token1",
  [anon_sym_BQUOTE] = "`",
  [aux_sym_string_token2] = "string_token2",
  [sym_escape_sequence] = "escape_sequence",
  [sym_operator] = "operator",
  [sym_punctuation] = "punctuation",
  [sym_source_file] = "source_file",
  [sym__statement] = "_statement",
  [sym_function_definition] = "function_definition",
  [sym_parameter_list] = "parameter_list",
  [sym_block] = "block",
  [sym_literal] = "literal",
  [sym_string] = "string",
  [aux_sym_source_file_repeat1] = "source_file_repeat1",
  [aux_sym_parameter_list_repeat1] = "parameter_list_repeat1",
  [aux_sym_string_repeat1] = "string_repeat1",
  [aux_sym_string_repeat2] = "string_repeat2",
};

static const TSSymbol ts_symbol_map[] = {
  [ts_builtin_sym_end] = ts_builtin_sym_end,
  [sym_comment] = sym_comment,
  [anon_sym_fn] = anon_sym_fn,
  [sym_keyword] = sym_keyword,
  [sym_identifier] = sym_identifier,
  [anon_sym_LPAREN] = anon_sym_LPAREN,
  [anon_sym_COMMA] = anon_sym_COMMA,
  [anon_sym_RPAREN] = anon_sym_RPAREN,
  [anon_sym_LBRACE] = anon_sym_LBRACE,
  [anon_sym_RBRACE] = anon_sym_RBRACE,
  [sym_boolean] = sym_boolean,
  [sym_null] = sym_null,
  [sym_number] = sym_number,
  [anon_sym_DQUOTE] = anon_sym_DQUOTE,
  [aux_sym_string_token1] = aux_sym_string_token1,
  [anon_sym_BQUOTE] = anon_sym_BQUOTE,
  [aux_sym_string_token2] = aux_sym_string_token2,
  [sym_escape_sequence] = sym_escape_sequence,
  [sym_operator] = sym_operator,
  [sym_punctuation] = sym_punctuation,
  [sym_source_file] = sym_source_file,
  [sym__statement] = sym__statement,
  [sym_function_definition] = sym_function_definition,
  [sym_parameter_list] = sym_parameter_list,
  [sym_block] = sym_block,
  [sym_literal] = sym_literal,
  [sym_string] = sym_string,
  [aux_sym_source_file_repeat1] = aux_sym_source_file_repeat1,
  [aux_sym_parameter_list_repeat1] = aux_sym_parameter_list_repeat1,
  [aux_sym_string_repeat1] = aux_sym_string_repeat1,
  [aux_sym_string_repeat2] = aux_sym_string_repeat2,
};

static const TSSymbolMetadata ts_symbol_metadata[] = {
  [ts_builtin_sym_end] = {
    .visible = false,
    .named = true,
  },
  [sym_comment] = {
    .visible = true,
    .named = true,
  },
  [anon_sym_fn] = {
    .visible = true,
    .named = false,
  },
  [sym_keyword] = {
    .visible = true,
    .named = true,
  },
  [sym_identifier] = {
    .visible = true,
    .named = true,
  },
  [anon_sym_LPAREN] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_COMMA] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_RPAREN] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_LBRACE] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_RBRACE] = {
    .visible = true,
    .named = false,
  },
  [sym_boolean] = {
    .visible = true,
    .named = true,
  },
  [sym_null] = {
    .visible = true,
    .named = true,
  },
  [sym_number] = {
    .visible = true,
    .named = true,
  },
  [anon_sym_DQUOTE] = {
    .visible = true,
    .named = false,
  },
  [aux_sym_string_token1] = {
    .visible = false,
    .named = false,
  },
  [anon_sym_BQUOTE] = {
    .visible = true,
    .named = false,
  },
  [aux_sym_string_token2] = {
    .visible = false,
    .named = false,
  },
  [sym_escape_sequence] = {
    .visible = true,
    .named = true,
  },
  [sym_operator] = {
    .visible = true,
    .named = true,
  },
  [sym_punctuation] = {
    .visible = true,
    .named = true,
  },
  [sym_source_file] = {
    .visible = true,
    .named = true,
  },
  [sym__statement] = {
    .visible = false,
    .named = true,
  },
  [sym_function_definition] = {
    .visible = true,
    .named = true,
  },
  [sym_parameter_list] = {
    .visible = true,
    .named = true,
  },
  [sym_block] = {
    .visible = true,
    .named = true,
  },
  [sym_literal] = {
    .visible = true,
    .named = true,
  },
  [sym_string] = {
    .visible = true,
    .named = true,
  },
  [aux_sym_source_file_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_parameter_list_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_string_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_string_repeat2] = {
    .visible = false,
    .named = false,
  },
};

static const TSSymbol ts_alias_sequences[PRODUCTION_ID_COUNT][MAX_ALIAS_SEQUENCE_LENGTH] = {
  [0] = {0},
};

static const uint16_t ts_non_terminal_alias_map[] = {
  0,
};

static const TSStateId ts_primary_state_ids[STATE_COUNT] = {
  [0] = 0,
  [1] = 1,
  [2] = 2,
  [3] = 3,
  [4] = 3,
  [5] = 2,
  [6] = 6,
  [7] = 7,
  [8] = 6,
  [9] = 9,
  [10] = 10,
  [11] = 11,
  [12] = 12,
  [13] = 11,
  [14] = 10,
  [15] = 9,
  [16] = 12,
  [17] = 17,
  [18] = 18,
  [19] = 18,
  [20] = 17,
  [21] = 21,
  [22] = 22,
  [23] = 23,
  [24] = 24,
  [25] = 21,
  [26] = 22,
  [27] = 24,
  [28] = 28,
  [29] = 29,
  [30] = 23,
  [31] = 31,
  [32] = 32,
  [33] = 33,
  [34] = 34,
  [35] = 35,
  [36] = 36,
  [37] = 37,
  [38] = 36,
  [39] = 35,
  [40] = 40,
  [41] = 41,
  [42] = 42,
  [43] = 43,
  [44] = 44,
  [45] = 45,
  [46] = 45,
};

static TSCharacterRange sym_escape_sequence_character_set_1[] = {
  {'"', '"'}, {'\'', '\''}, {'0', '7'}, {'U', 'U'}, {'\\', '\\'}, {'b', 'b'}, {'f', 'f'}, {'n', 'n'},
  {'r', 'r'}, {'t', 'v'}, {'x', 'x'},
};

static bool ts_lex(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      if (eof) ADVANCE(26);
      ADVANCE_MAP(
        '!', 87,
        '"', 75,
        '#', 27,
        '%', 87,
        '&', 8,
        '(', 62,
        ')', 64,
        '*', 6,
        '+', 89,
        ',', 63,
        '-', 85,
        '.', 92,
        '/', 87,
        '0', 68,
        '<', 86,
        '=', 87,
        '>', 88,
        '\\', 7,
        '^', 87,
        '`', 79,
        'b', 50,
        'c', 48,
        'e', 41,
        'f', 33,
        'i', 38,
        'n', 57,
        'r', 36,
        't', 32,
        '{', 65,
        '|', 90,
        '}', 66,
        ':', 91,
        ';', 91,
        '[', 91,
        ']', 91,
      );
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(0);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(69);
      if (('A' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 1:
      ADVANCE_MAP(
        '!', 87,
        '"', 75,
        '#', 27,
        '%', 87,
        '&', 8,
        '*', 6,
        '+', 89,
        '-', 85,
        '.', 92,
        '/', 87,
        '0', 68,
        '<', 86,
        '=', 87,
        '>', 88,
        '^', 87,
        '`', 79,
        'b', 50,
        'c', 48,
        'e', 41,
        'f', 33,
        'i', 38,
        'n', 57,
        'r', 36,
        't', 32,
        '|', 90,
        '}', 66,
      );
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(1);
      if (('(' <= lookahead && lookahead <= ',') ||
          lookahead == ':' ||
          lookahead == ';' ||
          lookahead == '[' ||
          lookahead == ']' ||
          lookahead == '{') ADVANCE(91);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(69);
      if (('A' <= lookahead && lookahead <= 'Z') ||
          ('_' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 2:
      if (lookahead == '"') ADVANCE(75);
      if (lookahead == '#') ADVANCE(76);
      if (lookahead == '\\') ADVANCE(7);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(77);
      if (lookahead != 0) ADVANCE(78);
      END_STATE();
    case 3:
      if (lookahead == '#') ADVANCE(27);
      if (lookahead == '(') ADVANCE(62);
      if (lookahead == ')') ADVANCE(64);
      if (lookahead == ',') ADVANCE(63);
      if (lookahead == '{') ADVANCE(65);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(3);
      if (('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 4:
      if (lookahead == '#') ADVANCE(80);
      if (lookahead == '\\') ADVANCE(7);
      if (lookahead == '`') ADVANCE(79);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(81);
      if (lookahead != 0) ADVANCE(82);
      END_STATE();
    case 5:
      if (lookahead == ')') ADVANCE(30);
      END_STATE();
    case 6:
      if (lookahead == '=') ADVANCE(84);
      END_STATE();
    case 7:
      if (lookahead == 'U') ADVANCE(24);
      if (lookahead == 'u') ADVANCE(20);
      if (lookahead == 'x') ADVANCE(18);
      if (('0' <= lookahead && lookahead <= '7')) ADVANCE(13);
      if (set_contains(sym_escape_sequence_character_set_1, 11, lookahead)) ADVANCE(83);
      END_STATE();
    case 8:
      if (lookahead == '^') ADVANCE(87);
      if (lookahead == '&' ||
          lookahead == '=') ADVANCE(84);
      END_STATE();
    case 9:
      if (lookahead == '+' ||
          lookahead == '-') ADVANCE(15);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(73);
      END_STATE();
    case 10:
      if (lookahead == '0' ||
          lookahead == '1') ADVANCE(71);
      END_STATE();
    case 11:
      if (('0' <= lookahead && lookahead <= '7')) ADVANCE(83);
      END_STATE();
    case 12:
      if (('0' <= lookahead && lookahead <= '7')) ADVANCE(72);
      END_STATE();
    case 13:
      if (('0' <= lookahead && lookahead <= '7')) ADVANCE(11);
      END_STATE();
    case 14:
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(70);
      END_STATE();
    case 15:
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(73);
      END_STATE();
    case 16:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(83);
      END_STATE();
    case 17:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(74);
      END_STATE();
    case 18:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(16);
      END_STATE();
    case 19:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(18);
      END_STATE();
    case 20:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(19);
      END_STATE();
    case 21:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(20);
      END_STATE();
    case 22:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(21);
      END_STATE();
    case 23:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(22);
      END_STATE();
    case 24:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(23);
      END_STATE();
    case 25:
      if (eof) ADVANCE(26);
      ADVANCE_MAP(
        '!', 87,
        '"', 75,
        '#', 27,
        '%', 87,
        '&', 8,
        '*', 6,
        '+', 89,
        '-', 85,
        '.', 92,
        '/', 87,
        '0', 68,
        '<', 86,
        '=', 87,
        '>', 88,
        '^', 87,
        '`', 79,
        'b', 50,
        'c', 48,
        'e', 41,
        'f', 33,
        'i', 38,
        'n', 57,
        'r', 36,
        't', 32,
        '|', 90,
      );
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(25);
      if (('(' <= lookahead && lookahead <= ',') ||
          lookahead == ':' ||
          lookahead == ';' ||
          lookahead == '[' ||
          lookahead == ']' ||
          ('{' <= lookahead && lookahead <= '}')) ADVANCE(91);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(69);
      if (('A' <= lookahead && lookahead <= 'Z') ||
          ('_' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 26:
      ACCEPT_TOKEN(ts_builtin_sym_end);
      END_STATE();
    case 27:
      ACCEPT_TOKEN(sym_comment);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(27);
      END_STATE();
    case 28:
      ACCEPT_TOKEN(anon_sym_fn);
      if (lookahead == '(') ADVANCE(5);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 29:
      ACCEPT_TOKEN(sym_keyword);
      if (lookahead == '(') ADVANCE(5);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 30:
      ACCEPT_TOKEN(sym_identifier);
      END_STATE();
    case 31:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'a') ADVANCE(40);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('b' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 32:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'a') ADVANCE(56);
      if (lookahead == 'r') ADVANCE(60);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('b' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 33:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'a') ADVANCE(44);
      if (lookahead == 'n') ADVANCE(28);
      if (lookahead == 'o') ADVANCE(49);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('b' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 34:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'e') ADVANCE(61);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 35:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'e') ADVANCE(29);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 36:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'e') ADVANCE(55);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 37:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'e') ADVANCE(31);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 38:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'f') ADVANCE(29);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 39:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'i') ADVANCE(47);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 40:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'k') ADVANCE(29);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 41:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'l') ADVANCE(52);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 42:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'l') ADVANCE(67);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 43:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'l') ADVANCE(42);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 44:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'l') ADVANCE(53);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 45:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'n') ADVANCE(29);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 46:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'n') ADVANCE(54);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 47:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'n') ADVANCE(59);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 48:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'o') ADVANCE(46);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 49:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'r') ADVANCE(29);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 50:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'r') ADVANCE(37);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 51:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'r') ADVANCE(45);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 52:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 's') ADVANCE(35);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 53:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 's') ADVANCE(34);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 54:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 't') ADVANCE(39);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 55:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 't') ADVANCE(58);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 56:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'u') ADVANCE(29);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 57:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'u') ADVANCE(43);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 58:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'u') ADVANCE(51);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 59:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'u') ADVANCE(35);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 60:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (lookahead == 'u') ADVANCE(34);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 61:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '(') ADVANCE(5);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 62:
      ACCEPT_TOKEN(anon_sym_LPAREN);
      END_STATE();
    case 63:
      ACCEPT_TOKEN(anon_sym_COMMA);
      END_STATE();
    case 64:
      ACCEPT_TOKEN(anon_sym_RPAREN);
      END_STATE();
    case 65:
      ACCEPT_TOKEN(anon_sym_LBRACE);
      END_STATE();
    case 66:
      ACCEPT_TOKEN(anon_sym_RBRACE);
      END_STATE();
    case 67:
      ACCEPT_TOKEN(sym_null);
      if (lookahead == '(') ADVANCE(5);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(61);
      END_STATE();
    case 68:
      ACCEPT_TOKEN(sym_number);
      if (lookahead == '.') ADVANCE(14);
      if (lookahead == 'b') ADVANCE(10);
      if (lookahead == 'o') ADVANCE(12);
      if (lookahead == 'x') ADVANCE(17);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(9);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(69);
      END_STATE();
    case 69:
      ACCEPT_TOKEN(sym_number);
      if (lookahead == '.') ADVANCE(14);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(9);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(69);
      END_STATE();
    case 70:
      ACCEPT_TOKEN(sym_number);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(9);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(70);
      END_STATE();
    case 71:
      ACCEPT_TOKEN(sym_number);
      if (lookahead == '0' ||
          lookahead == '1') ADVANCE(71);
      END_STATE();
    case 72:
      ACCEPT_TOKEN(sym_number);
      if (('0' <= lookahead && lookahead <= '7')) ADVANCE(72);
      END_STATE();
    case 73:
      ACCEPT_TOKEN(sym_number);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(73);
      END_STATE();
    case 74:
      ACCEPT_TOKEN(sym_number);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(74);
      END_STATE();
    case 75:
      ACCEPT_TOKEN(anon_sym_DQUOTE);
      END_STATE();
    case 76:
      ACCEPT_TOKEN(aux_sym_string_token1);
      if (lookahead == '\n') ADVANCE(78);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '\\') ADVANCE(76);
      END_STATE();
    case 77:
      ACCEPT_TOKEN(aux_sym_string_token1);
      if (lookahead == '#') ADVANCE(76);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(77);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '#' &&
          lookahead != '\\') ADVANCE(78);
      END_STATE();
    case 78:
      ACCEPT_TOKEN(aux_sym_string_token1);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '\\') ADVANCE(78);
      END_STATE();
    case 79:
      ACCEPT_TOKEN(anon_sym_BQUOTE);
      END_STATE();
    case 80:
      ACCEPT_TOKEN(aux_sym_string_token2);
      if (lookahead == '\n') ADVANCE(82);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(80);
      END_STATE();
    case 81:
      ACCEPT_TOKEN(aux_sym_string_token2);
      if (lookahead == '#') ADVANCE(80);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(81);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(82);
      END_STATE();
    case 82:
      ACCEPT_TOKEN(aux_sym_string_token2);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(82);
      END_STATE();
    case 83:
      ACCEPT_TOKEN(sym_escape_sequence);
      END_STATE();
    case 84:
      ACCEPT_TOKEN(sym_operator);
      END_STATE();
    case 85:
      ACCEPT_TOKEN(sym_operator);
      if (lookahead == '.') ADVANCE(14);
      if (lookahead == '-' ||
          lookahead == '=') ADVANCE(84);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(69);
      END_STATE();
    case 86:
      ACCEPT_TOKEN(sym_operator);
      if (lookahead == '<') ADVANCE(87);
      if (lookahead == '=') ADVANCE(84);
      END_STATE();
    case 87:
      ACCEPT_TOKEN(sym_operator);
      if (lookahead == '=') ADVANCE(84);
      END_STATE();
    case 88:
      ACCEPT_TOKEN(sym_operator);
      if (lookahead == '=') ADVANCE(84);
      if (lookahead == '>') ADVANCE(87);
      END_STATE();
    case 89:
      ACCEPT_TOKEN(sym_operator);
      if (lookahead == '+' ||
          lookahead == '=') ADVANCE(84);
      END_STATE();
    case 90:
      ACCEPT_TOKEN(sym_operator);
      if (lookahead == '=' ||
          lookahead == '|') ADVANCE(84);
      END_STATE();
    case 91:
      ACCEPT_TOKEN(sym_punctuation);
      END_STATE();
    case 92:
      ACCEPT_TOKEN(sym_punctuation);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(70);
      END_STATE();
    default:
      return false;
  }
}

static const TSLexMode ts_lex_modes[STATE_COUNT] = {
  [0] = {.lex_state = 0},
  [1] = {.lex_state = 25},
  [2] = {.lex_state = 1},
  [3] = {.lex_state = 1},
  [4] = {.lex_state = 1},
  [5] = {.lex_state = 1},
  [6] = {.lex_state = 1},
  [7] = {.lex_state = 25},
  [8] = {.lex_state = 25},
  [9] = {.lex_state = 1},
  [10] = {.lex_state = 25},
  [11] = {.lex_state = 25},
  [12] = {.lex_state = 25},
  [13] = {.lex_state = 1},
  [14] = {.lex_state = 1},
  [15] = {.lex_state = 25},
  [16] = {.lex_state = 1},
  [17] = {.lex_state = 25},
  [18] = {.lex_state = 1},
  [19] = {.lex_state = 25},
  [20] = {.lex_state = 1},
  [21] = {.lex_state = 4},
  [22] = {.lex_state = 2},
  [23] = {.lex_state = 2},
  [24] = {.lex_state = 4},
  [25] = {.lex_state = 4},
  [26] = {.lex_state = 2},
  [27] = {.lex_state = 4},
  [28] = {.lex_state = 4},
  [29] = {.lex_state = 2},
  [30] = {.lex_state = 2},
  [31] = {.lex_state = 3},
  [32] = {.lex_state = 3},
  [33] = {.lex_state = 3},
  [34] = {.lex_state = 3},
  [35] = {.lex_state = 3},
  [36] = {.lex_state = 3},
  [37] = {.lex_state = 3},
  [38] = {.lex_state = 3},
  [39] = {.lex_state = 3},
  [40] = {.lex_state = 3},
  [41] = {.lex_state = 0},
  [42] = {.lex_state = 3},
  [43] = {.lex_state = 3},
  [44] = {.lex_state = 3},
  [45] = {.lex_state = 3},
  [46] = {.lex_state = 3},
};

static const uint16_t ts_parse_table[LARGE_STATE_COUNT][SYMBOL_COUNT] = {
  [0] = {
    [ts_builtin_sym_end] = ACTIONS(1),
    [sym_comment] = ACTIONS(3),
    [anon_sym_fn] = ACTIONS(1),
    [sym_keyword] = ACTIONS(1),
    [sym_identifier] = ACTIONS(1),
    [anon_sym_LPAREN] = ACTIONS(1),
    [anon_sym_COMMA] = ACTIONS(1),
    [anon_sym_RPAREN] = ACTIONS(1),
    [anon_sym_LBRACE] = ACTIONS(1),
    [anon_sym_RBRACE] = ACTIONS(1),
    [sym_boolean] = ACTIONS(1),
    [sym_null] = ACTIONS(1),
    [sym_number] = ACTIONS(1),
    [anon_sym_DQUOTE] = ACTIONS(1),
    [anon_sym_BQUOTE] = ACTIONS(1),
    [sym_escape_sequence] = ACTIONS(1),
    [sym_operator] = ACTIONS(1),
    [sym_punctuation] = ACTIONS(1),
  },
  [1] = {
    [sym_source_file] = STATE(41),
    [sym__statement] = STATE(7),
    [sym_function_definition] = STATE(7),
    [sym_literal] = STATE(7),
    [sym_string] = STATE(12),
    [aux_sym_source_file_repeat1] = STATE(7),
    [ts_builtin_sym_end] = ACTIONS(5),
    [sym_comment] = ACTIONS(3),
    [anon_sym_fn] = ACTIONS(7),
    [sym_keyword] = ACTIONS(9),
    [sym_identifier] = ACTIONS(9),
    [sym_boolean] = ACTIONS(11),
    [sym_null] = ACTIONS(11),
    [sym_number] = ACTIONS(13),
    [anon_sym_DQUOTE] = ACTIONS(15),
    [anon_sym_BQUOTE] = ACTIONS(17),
    [sym_operator] = ACTIONS(9),
    [sym_punctuation] = ACTIONS(9),
  },
  [2] = {
    [sym__statement] = STATE(6),
    [sym_function_definition] = STATE(6),
    [sym_literal] = STATE(6),
    [sym_string] = STATE(16),
    [aux_sym_source_file_repeat1] = STATE(6),
    [sym_comment] = ACTIONS(3),
    [anon_sym_fn] = ACTIONS(19),
    [sym_keyword] = ACTIONS(21),
    [sym_identifier] = ACTIONS(21),
    [anon_sym_RBRACE] = ACTIONS(23),
    [sym_boolean] = ACTIONS(25),
    [sym_null] = ACTIONS(25),
    [sym_number] = ACTIONS(27),
    [anon_sym_DQUOTE] = ACTIONS(29),
    [anon_sym_BQUOTE] = ACTIONS(31),
    [sym_operator] = ACTIONS(21),
    [sym_punctuation] = ACTIONS(21),
  },
  [3] = {
    [sym__statement] = STATE(5),
    [sym_function_definition] = STATE(5),
    [sym_literal] = STATE(5),
    [sym_string] = STATE(16),
    [aux_sym_source_file_repeat1] = STATE(5),
    [sym_comment] = ACTIONS(3),
    [anon_sym_fn] = ACTIONS(19),
    [sym_keyword] = ACTIONS(33),
    [sym_identifier] = ACTIONS(33),
    [anon_sym_RBRACE] = ACTIONS(35),
    [sym_boolean] = ACTIONS(25),
    [sym_null] = ACTIONS(25),
    [sym_number] = ACTIONS(27),
    [anon_sym_DQUOTE] = ACTIONS(29),
    [anon_sym_BQUOTE] = ACTIONS(31),
    [sym_operator] = ACTIONS(33),
    [sym_punctuation] = ACTIONS(33),
  },
  [4] = {
    [sym__statement] = STATE(2),
    [sym_function_definition] = STATE(2),
    [sym_literal] = STATE(2),
    [sym_string] = STATE(16),
    [aux_sym_source_file_repeat1] = STATE(2),
    [sym_comment] = ACTIONS(3),
    [anon_sym_fn] = ACTIONS(19),
    [sym_keyword] = ACTIONS(37),
    [sym_identifier] = ACTIONS(37),
    [anon_sym_RBRACE] = ACTIONS(39),
    [sym_boolean] = ACTIONS(25),
    [sym_null] = ACTIONS(25),
    [sym_number] = ACTIONS(27),
    [anon_sym_DQUOTE] = ACTIONS(29),
    [anon_sym_BQUOTE] = ACTIONS(31),
    [sym_operator] = ACTIONS(37),
    [sym_punctuation] = ACTIONS(37),
  },
  [5] = {
    [sym__statement] = STATE(6),
    [sym_function_definition] = STATE(6),
    [sym_literal] = STATE(6),
    [sym_string] = STATE(16),
    [aux_sym_source_file_repeat1] = STATE(6),
    [sym_comment] = ACTIONS(3),
    [anon_sym_fn] = ACTIONS(19),
    [sym_keyword] = ACTIONS(21),
    [sym_identifier] = ACTIONS(21),
    [anon_sym_RBRACE] = ACTIONS(41),
    [sym_boolean] = ACTIONS(25),
    [sym_null] = ACTIONS(25),
    [sym_number] = ACTIONS(27),
    [anon_sym_DQUOTE] = ACTIONS(29),
    [anon_sym_BQUOTE] = ACTIONS(31),
    [sym_operator] = ACTIONS(21),
    [sym_punctuation] = ACTIONS(21),
  },
  [6] = {
    [sym__statement] = STATE(6),
    [sym_function_definition] = STATE(6),
    [sym_literal] = STATE(6),
    [sym_string] = STATE(16),
    [aux_sym_source_file_repeat1] = STATE(6),
    [sym_comment] = ACTIONS(3),
    [anon_sym_fn] = ACTIONS(43),
    [sym_keyword] = ACTIONS(46),
    [sym_identifier] = ACTIONS(46),
    [anon_sym_RBRACE] = ACTIONS(49),
    [sym_boolean] = ACTIONS(51),
    [sym_null] = ACTIONS(51),
    [sym_number] = ACTIONS(54),
    [anon_sym_DQUOTE] = ACTIONS(57),
    [anon_sym_BQUOTE] = ACTIONS(60),
    [sym_operator] = ACTIONS(46),
    [sym_punctuation] = ACTIONS(46),
  },
  [7] = {
    [sym__statement] = STATE(8),
    [sym_function_definition] = STATE(8),
    [sym_literal] = STATE(8),
    [sym_string] = STATE(12),
    [aux_sym_source_file_repeat1] = STATE(8),
    [ts_builtin_sym_end] = ACTIONS(63),
    [sym_comment] = ACTIONS(3),
    [anon_sym_fn] = ACTIONS(7),
    [sym_keyword] = ACTIONS(65),
    [sym_identifier] = ACTIONS(65),
    [sym_boolean] = ACTIONS(11),
    [sym_null] = ACTIONS(11),
    [sym_number] = ACTIONS(13),
    [anon_sym_DQUOTE] = ACTIONS(15),
    [anon_sym_BQUOTE] = ACTIONS(17),
    [sym_operator] = ACTIONS(65),
    [sym_punctuation] = ACTIONS(65),
  },
  [8] = {
    [sym__statement] = STATE(8),
    [sym_function_definition] = STATE(8),
    [sym_literal] = STATE(8),
    [sym_string] = STATE(12),
    [aux_sym_source_file_repeat1] = STATE(8),
    [ts_builtin_sym_end] = ACTIONS(49),
    [sym_comment] = ACTIONS(3),
    [anon_sym_fn] = ACTIONS(67),
    [sym_keyword] = ACTIONS(70),
    [sym_identifier] = ACTIONS(70),
    [sym_boolean] = ACTIONS(73),
    [sym_null] = ACTIONS(73),
    [sym_number] = ACTIONS(76),
    [anon_sym_DQUOTE] = ACTIONS(79),
    [anon_sym_BQUOTE] = ACTIONS(82),
    [sym_operator] = ACTIONS(70),
    [sym_punctuation] = ACTIONS(70),
  },
};

static const uint16_t ts_small_parse_table[] = {
  [0] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(87), 4,
      anon_sym_RBRACE,
      sym_number,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
    ACTIONS(85), 7,
      anon_sym_fn,
      sym_keyword,
      sym_identifier,
      sym_boolean,
      sym_null,
      sym_operator,
      sym_punctuation,
  [19] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(89), 4,
      ts_builtin_sym_end,
      sym_number,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
    ACTIONS(91), 7,
      anon_sym_fn,
      sym_keyword,
      sym_identifier,
      sym_boolean,
      sym_null,
      sym_operator,
      sym_punctuation,
  [38] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(93), 4,
      ts_builtin_sym_end,
      sym_number,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
    ACTIONS(95), 7,
      anon_sym_fn,
      sym_keyword,
      sym_identifier,
      sym_boolean,
      sym_null,
      sym_operator,
      sym_punctuation,
  [57] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(97), 4,
      ts_builtin_sym_end,
      sym_number,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
    ACTIONS(99), 7,
      anon_sym_fn,
      sym_keyword,
      sym_identifier,
      sym_boolean,
      sym_null,
      sym_operator,
      sym_punctuation,
  [76] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(93), 4,
      anon_sym_RBRACE,
      sym_number,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
    ACTIONS(95), 7,
      anon_sym_fn,
      sym_keyword,
      sym_identifier,
      sym_boolean,
      sym_null,
      sym_operator,
      sym_punctuation,
  [95] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(89), 4,
      anon_sym_RBRACE,
      sym_number,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
    ACTIONS(91), 7,
      anon_sym_fn,
      sym_keyword,
      sym_identifier,
      sym_boolean,
      sym_null,
      sym_operator,
      sym_punctuation,
  [114] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(87), 4,
      ts_builtin_sym_end,
      sym_number,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
    ACTIONS(85), 7,
      anon_sym_fn,
      sym_keyword,
      sym_identifier,
      sym_boolean,
      sym_null,
      sym_operator,
      sym_punctuation,
  [133] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(97), 4,
      anon_sym_RBRACE,
      sym_number,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
    ACTIONS(99), 7,
      anon_sym_fn,
      sym_keyword,
      sym_identifier,
      sym_boolean,
      sym_null,
      sym_operator,
      sym_punctuation,
  [152] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(101), 4,
      ts_builtin_sym_end,
      sym_number,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
    ACTIONS(103), 7,
      anon_sym_fn,
      sym_keyword,
      sym_identifier,
      sym_boolean,
      sym_null,
      sym_operator,
      sym_punctuation,
  [171] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(107), 4,
      anon_sym_RBRACE,
      sym_number,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
    ACTIONS(105), 7,
      anon_sym_fn,
      sym_keyword,
      sym_identifier,
      sym_boolean,
      sym_null,
      sym_operator,
      sym_punctuation,
  [190] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(107), 4,
      ts_builtin_sym_end,
      sym_number,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
    ACTIONS(105), 7,
      anon_sym_fn,
      sym_keyword,
      sym_identifier,
      sym_boolean,
      sym_null,
      sym_operator,
      sym_punctuation,
  [209] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(101), 4,
      anon_sym_RBRACE,
      sym_number,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
    ACTIONS(103), 7,
      anon_sym_fn,
      sym_keyword,
      sym_identifier,
      sym_boolean,
      sym_null,
      sym_operator,
      sym_punctuation,
  [228] = 5,
    ACTIONS(109), 1,
      sym_comment,
    ACTIONS(111), 1,
      anon_sym_BQUOTE,
    ACTIONS(113), 1,
      aux_sym_string_token2,
    ACTIONS(115), 1,
      sym_escape_sequence,
    STATE(28), 1,
      aux_sym_string_repeat2,
  [244] = 5,
    ACTIONS(109), 1,
      sym_comment,
    ACTIONS(111), 1,
      anon_sym_DQUOTE,
    ACTIONS(117), 1,
      aux_sym_string_token1,
    ACTIONS(119), 1,
      sym_escape_sequence,
    STATE(29), 1,
      aux_sym_string_repeat1,
  [260] = 5,
    ACTIONS(109), 1,
      sym_comment,
    ACTIONS(121), 1,
      anon_sym_DQUOTE,
    ACTIONS(123), 1,
      aux_sym_string_token1,
    ACTIONS(125), 1,
      sym_escape_sequence,
    STATE(22), 1,
      aux_sym_string_repeat1,
  [276] = 5,
    ACTIONS(109), 1,
      sym_comment,
    ACTIONS(121), 1,
      anon_sym_BQUOTE,
    ACTIONS(127), 1,
      aux_sym_string_token2,
    ACTIONS(129), 1,
      sym_escape_sequence,
    STATE(21), 1,
      aux_sym_string_repeat2,
  [292] = 5,
    ACTIONS(109), 1,
      sym_comment,
    ACTIONS(113), 1,
      aux_sym_string_token2,
    ACTIONS(115), 1,
      sym_escape_sequence,
    ACTIONS(131), 1,
      anon_sym_BQUOTE,
    STATE(28), 1,
      aux_sym_string_repeat2,
  [308] = 5,
    ACTIONS(109), 1,
      sym_comment,
    ACTIONS(117), 1,
      aux_sym_string_token1,
    ACTIONS(119), 1,
      sym_escape_sequence,
    ACTIONS(131), 1,
      anon_sym_DQUOTE,
    STATE(29), 1,
      aux_sym_string_repeat1,
  [324] = 5,
    ACTIONS(109), 1,
      sym_comment,
    ACTIONS(133), 1,
      anon_sym_BQUOTE,
    ACTIONS(135), 1,
      aux_sym_string_token2,
    ACTIONS(137), 1,
      sym_escape_sequence,
    STATE(25), 1,
      aux_sym_string_repeat2,
  [340] = 5,
    ACTIONS(109), 1,
      sym_comment,
    ACTIONS(139), 1,
      anon_sym_BQUOTE,
    ACTIONS(141), 1,
      aux_sym_string_token2,
    ACTIONS(144), 1,
      sym_escape_sequence,
    STATE(28), 1,
      aux_sym_string_repeat2,
  [356] = 5,
    ACTIONS(109), 1,
      sym_comment,
    ACTIONS(147), 1,
      anon_sym_DQUOTE,
    ACTIONS(149), 1,
      aux_sym_string_token1,
    ACTIONS(152), 1,
      sym_escape_sequence,
    STATE(29), 1,
      aux_sym_string_repeat1,
  [372] = 5,
    ACTIONS(109), 1,
      sym_comment,
    ACTIONS(133), 1,
      anon_sym_DQUOTE,
    ACTIONS(155), 1,
      aux_sym_string_token1,
    ACTIONS(157), 1,
      sym_escape_sequence,
    STATE(26), 1,
      aux_sym_string_repeat1,
  [388] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(159), 1,
      anon_sym_COMMA,
    ACTIONS(161), 1,
      anon_sym_RPAREN,
    STATE(32), 1,
      aux_sym_parameter_list_repeat1,
  [401] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(163), 1,
      anon_sym_COMMA,
    ACTIONS(166), 1,
      anon_sym_RPAREN,
    STATE(32), 1,
      aux_sym_parameter_list_repeat1,
  [414] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(159), 1,
      anon_sym_COMMA,
    ACTIONS(168), 1,
      anon_sym_RPAREN,
    STATE(31), 1,
      aux_sym_parameter_list_repeat1,
  [427] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(170), 1,
      sym_identifier,
    ACTIONS(172), 1,
      anon_sym_RPAREN,
  [437] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(174), 1,
      anon_sym_LPAREN,
    STATE(36), 1,
      sym_parameter_list,
  [447] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(176), 1,
      anon_sym_LBRACE,
    STATE(11), 1,
      sym_block,
  [457] = 2,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(166), 2,
      anon_sym_COMMA,
      anon_sym_RPAREN,
  [465] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(178), 1,
      anon_sym_LBRACE,
    STATE(13), 1,
      sym_block,
  [475] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(174), 1,
      anon_sym_LPAREN,
    STATE(38), 1,
      sym_parameter_list,
  [485] = 2,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(180), 1,
      sym_identifier,
  [492] = 2,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(182), 1,
      ts_builtin_sym_end,
  [499] = 2,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(184), 1,
      anon_sym_LBRACE,
  [506] = 2,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(186), 1,
      anon_sym_LBRACE,
  [513] = 2,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(188), 1,
      anon_sym_LBRACE,
  [520] = 2,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(190), 1,
      sym_identifier,
  [527] = 2,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(192), 1,
      sym_identifier,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(9)] = 0,
  [SMALL_STATE(10)] = 19,
  [SMALL_STATE(11)] = 38,
  [SMALL_STATE(12)] = 57,
  [SMALL_STATE(13)] = 76,
  [SMALL_STATE(14)] = 95,
  [SMALL_STATE(15)] = 114,
  [SMALL_STATE(16)] = 133,
  [SMALL_STATE(17)] = 152,
  [SMALL_STATE(18)] = 171,
  [SMALL_STATE(19)] = 190,
  [SMALL_STATE(20)] = 209,
  [SMALL_STATE(21)] = 228,
  [SMALL_STATE(22)] = 244,
  [SMALL_STATE(23)] = 260,
  [SMALL_STATE(24)] = 276,
  [SMALL_STATE(25)] = 292,
  [SMALL_STATE(26)] = 308,
  [SMALL_STATE(27)] = 324,
  [SMALL_STATE(28)] = 340,
  [SMALL_STATE(29)] = 356,
  [SMALL_STATE(30)] = 372,
  [SMALL_STATE(31)] = 388,
  [SMALL_STATE(32)] = 401,
  [SMALL_STATE(33)] = 414,
  [SMALL_STATE(34)] = 427,
  [SMALL_STATE(35)] = 437,
  [SMALL_STATE(36)] = 447,
  [SMALL_STATE(37)] = 457,
  [SMALL_STATE(38)] = 465,
  [SMALL_STATE(39)] = 475,
  [SMALL_STATE(40)] = 485,
  [SMALL_STATE(41)] = 492,
  [SMALL_STATE(42)] = 499,
  [SMALL_STATE(43)] = 506,
  [SMALL_STATE(44)] = 513,
  [SMALL_STATE(45)] = 520,
  [SMALL_STATE(46)] = 527,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = true}}, SHIFT_EXTRA(),
  [5] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 0, 0, 0),
  [7] = {.entry = {.count = 1, .reusable = false}}, SHIFT(45),
  [9] = {.entry = {.count = 1, .reusable = false}}, SHIFT(7),
  [11] = {.entry = {.count = 1, .reusable = false}}, SHIFT(12),
  [13] = {.entry = {.count = 1, .reusable = true}}, SHIFT(12),
  [15] = {.entry = {.count = 1, .reusable = true}}, SHIFT(23),
  [17] = {.entry = {.count = 1, .reusable = true}}, SHIFT(24),
  [19] = {.entry = {.count = 1, .reusable = false}}, SHIFT(46),
  [21] = {.entry = {.count = 1, .reusable = false}}, SHIFT(6),
  [23] = {.entry = {.count = 1, .reusable = true}}, SHIFT(20),
  [25] = {.entry = {.count = 1, .reusable = false}}, SHIFT(16),
  [27] = {.entry = {.count = 1, .reusable = true}}, SHIFT(16),
  [29] = {.entry = {.count = 1, .reusable = true}}, SHIFT(30),
  [31] = {.entry = {.count = 1, .reusable = true}}, SHIFT(27),
  [33] = {.entry = {.count = 1, .reusable = false}}, SHIFT(5),
  [35] = {.entry = {.count = 1, .reusable = true}}, SHIFT(19),
  [37] = {.entry = {.count = 1, .reusable = false}}, SHIFT(2),
  [39] = {.entry = {.count = 1, .reusable = true}}, SHIFT(18),
  [41] = {.entry = {.count = 1, .reusable = true}}, SHIFT(17),
  [43] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(46),
  [46] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(6),
  [49] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0),
  [51] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(16),
  [54] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(16),
  [57] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(30),
  [60] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(27),
  [63] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1, 0, 0),
  [65] = {.entry = {.count = 1, .reusable = false}}, SHIFT(8),
  [67] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(45),
  [70] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(8),
  [73] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(12),
  [76] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(12),
  [79] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(23),
  [82] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(24),
  [85] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 3, 0, 0),
  [87] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 3, 0, 0),
  [89] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 2, 0, 0),
  [91] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 2, 0, 0),
  [93] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_function_definition, 4, 0, 0),
  [95] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_function_definition, 4, 0, 0),
  [97] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_literal, 1, 0, 0),
  [99] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_literal, 1, 0, 0),
  [101] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_block, 3, 0, 0),
  [103] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_block, 3, 0, 0),
  [105] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_block, 2, 0, 0),
  [107] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_block, 2, 0, 0),
  [109] = {.entry = {.count = 1, .reusable = false}}, SHIFT_EXTRA(),
  [111] = {.entry = {.count = 1, .reusable = false}}, SHIFT(15),
  [113] = {.entry = {.count = 1, .reusable = true}}, SHIFT(28),
  [115] = {.entry = {.count = 1, .reusable = false}}, SHIFT(28),
  [117] = {.entry = {.count = 1, .reusable = true}}, SHIFT(29),
  [119] = {.entry = {.count = 1, .reusable = false}}, SHIFT(29),
  [121] = {.entry = {.count = 1, .reusable = false}}, SHIFT(10),
  [123] = {.entry = {.count = 1, .reusable = true}}, SHIFT(22),
  [125] = {.entry = {.count = 1, .reusable = false}}, SHIFT(22),
  [127] = {.entry = {.count = 1, .reusable = true}}, SHIFT(21),
  [129] = {.entry = {.count = 1, .reusable = false}}, SHIFT(21),
  [131] = {.entry = {.count = 1, .reusable = false}}, SHIFT(9),
  [133] = {.entry = {.count = 1, .reusable = false}}, SHIFT(14),
  [135] = {.entry = {.count = 1, .reusable = true}}, SHIFT(25),
  [137] = {.entry = {.count = 1, .reusable = false}}, SHIFT(25),
  [139] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0),
  [141] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0), SHIFT_REPEAT(28),
  [144] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0), SHIFT_REPEAT(28),
  [147] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0),
  [149] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0), SHIFT_REPEAT(29),
  [152] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0), SHIFT_REPEAT(29),
  [155] = {.entry = {.count = 1, .reusable = true}}, SHIFT(26),
  [157] = {.entry = {.count = 1, .reusable = false}}, SHIFT(26),
  [159] = {.entry = {.count = 1, .reusable = true}}, SHIFT(40),
  [161] = {.entry = {.count = 1, .reusable = true}}, SHIFT(43),
  [163] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_parameter_list_repeat1, 2, 0, 0), SHIFT_REPEAT(40),
  [166] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_parameter_list_repeat1, 2, 0, 0),
  [168] = {.entry = {.count = 1, .reusable = true}}, SHIFT(42),
  [170] = {.entry = {.count = 1, .reusable = true}}, SHIFT(33),
  [172] = {.entry = {.count = 1, .reusable = true}}, SHIFT(44),
  [174] = {.entry = {.count = 1, .reusable = true}}, SHIFT(34),
  [176] = {.entry = {.count = 1, .reusable = true}}, SHIFT(3),
  [178] = {.entry = {.count = 1, .reusable = true}}, SHIFT(4),
  [180] = {.entry = {.count = 1, .reusable = true}}, SHIFT(37),
  [182] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
  [184] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_parameter_list, 3, 0, 0),
  [186] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_parameter_list, 4, 0, 0),
  [188] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_parameter_list, 2, 0, 0),
  [190] = {.entry = {.count = 1, .reusable = true}}, SHIFT(35),
  [192] = {.entry = {.count = 1, .reusable = true}}, SHIFT(39),
};

#ifdef __cplusplus
extern "C" {
#endif
#ifdef TREE_SITTER_HIDE_SYMBOLS
#define TS_PUBLIC
#elif defined(_WIN32)
#define TS_PUBLIC __declspec(dllexport)
#else
#define TS_PUBLIC __attribute__((visibility("default")))
#endif

TS_PUBLIC const TSLanguage *tree_sitter_tau(void) {
  static const TSLanguage language = {
    .version = LANGUAGE_VERSION,
    .symbol_count = SYMBOL_COUNT,
    .alias_count = ALIAS_COUNT,
    .token_count = TOKEN_COUNT,
    .external_token_count = EXTERNAL_TOKEN_COUNT,
    .state_count = STATE_COUNT,
    .large_state_count = LARGE_STATE_COUNT,
    .production_id_count = PRODUCTION_ID_COUNT,
    .field_count = FIELD_COUNT,
    .max_alias_sequence_length = MAX_ALIAS_SEQUENCE_LENGTH,
    .parse_table = &ts_parse_table[0][0],
    .small_parse_table = ts_small_parse_table,
    .small_parse_table_map = ts_small_parse_table_map,
    .parse_actions = ts_parse_actions,
    .symbol_names = ts_symbol_names,
    .symbol_metadata = ts_symbol_metadata,
    .public_symbol_map = ts_symbol_map,
    .alias_map = ts_non_terminal_alias_map,
    .alias_sequences = &ts_alias_sequences[0][0],
    .lex_modes = ts_lex_modes,
    .lex_fn = ts_lex,
    .primary_state_ids = ts_primary_state_ids,
  };
  return &language;
}
#ifdef __cplusplus
}
#endif
