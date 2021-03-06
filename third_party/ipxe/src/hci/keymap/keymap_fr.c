/** @file
 *
 * "fr" keyboard mapping
 *
 * This file is automatically generated; do not edit
 *
 */

FILE_LICENCE ( PUBLIC_DOMAIN );

#include <ipxe/keymap.h>

/** "fr" basic remapping */
static struct keymap_key fr_basic[] = {
	{ 0x01, 0x11 },	/* Ctrl-A => Ctrl-Q */
	{ 0x11, 0x01 },	/* Ctrl-Q => Ctrl-A */
	{ 0x17, 0x1a },	/* Ctrl-W => Ctrl-Z */
	{ 0x1a, 0x17 },	/* Ctrl-Z => Ctrl-W */
	{ 0x1c, 0x2a },	/* 0x1c => '*' */
	{ 0x1d, 0x24 },	/* 0x1d => '$' */
	{ 0x1e, 0x1c },	/* 0x1e => 0x1c */
	{ 0x1f, 0x1d },	/* 0x1f => 0x1d */
	{ 0x21, 0x31 },	/* '!' => '1' */
	{ 0x22, 0x25 },	/* '"' => '%' */
	{ 0x23, 0x33 },	/* '#' => '3' */
	{ 0x24, 0x34 },	/* '$' => '4' */
	{ 0x25, 0x35 },	/* '%' => '5' */
	{ 0x26, 0x37 },	/* '&' => '7' */
	{ 0x28, 0x39 },	/* '(' => '9' */
	{ 0x29, 0x30 },	/* ')' => '0' */
	{ 0x2a, 0x38 },	/* '*' => '8' */
	{ 0x2c, 0x3b },	/* ',' => ';' */
	{ 0x2d, 0x29 },	/* '-' => ')' */
	{ 0x2e, 0x3a },	/* '.' => ':' */
	{ 0x2f, 0x21 },	/* '/' => '!' */
	{ 0x31, 0x26 },	/* '1' => '&' */
	{ 0x33, 0x22 },	/* '3' => '"' */
	{ 0x34, 0x27 },	/* '4' => '\'' */
	{ 0x35, 0x28 },	/* '5' => '(' */
	{ 0x36, 0x2d },	/* '6' => '-' */
	{ 0x38, 0x5f },	/* '8' => '_' */
	{ 0x3a, 0x4d },	/* ':' => 'M' */
	{ 0x3b, 0x6d },	/* ';' => 'm' */
	{ 0x3c, 0x2e },	/* '<' => '.' */
	{ 0x3e, 0x2f },	/* '>' => '/' */
	{ 0x40, 0x32 },	/* '@' => '2' */
	{ 0x41, 0x51 },	/* 'A' => 'Q' */
	{ 0x4d, 0x3f },	/* 'M' => '?' */
	{ 0x51, 0x41 },	/* 'Q' => 'A' */
	{ 0x57, 0x5a },	/* 'W' => 'Z' */
	{ 0x5a, 0x57 },	/* 'Z' => 'W' */
	{ 0x5b, 0x5e },	/* '[' => '^' */
	{ 0x5c, 0x2a },	/* '\\' => '*' */
	{ 0x5d, 0x24 },	/* ']' => '$' */
	{ 0x5e, 0x36 },	/* '^' => '6' */
	{ 0x61, 0x71 },	/* 'a' => 'q' */
	{ 0x6d, 0x2c },	/* 'm' => ',' */
	{ 0x71, 0x61 },	/* 'q' => 'a' */
	{ 0x77, 0x7a },	/* 'w' => 'z' */
	{ 0x7a, 0x77 },	/* 'z' => 'w' */
	{ 0xdc, 0x3c },	/* Pseudo-'\\' => '<' */
	{ 0xfc, 0x3e },	/* Pseudo-'|' => '>' */
	{ 0, 0 }
};

/** "fr" AltGr remapping */
static struct keymap_key fr_altgr[] = {
	{ 0x25, 0x5b },	/* '%' => '[' */
	{ 0x26, 0x60 },	/* '&' => '`' */
	{ 0x29, 0x40 },	/* ')' => '@' */
	{ 0x2a, 0x5c },	/* '*' => '\\' */
	{ 0x2b, 0x7d },	/* '+' => '}' */
	{ 0x2d, 0x5d },	/* '-' => ']' */
	{ 0x30, 0x40 },	/* '0' => '@' */
	{ 0x33, 0x23 },	/* '3' => '#' */
	{ 0x34, 0x7b },	/* '4' => '{' */
	{ 0x35, 0x5b },	/* '5' => '[' */
	{ 0x36, 0x7c },	/* '6' => '|' */
	{ 0x37, 0x60 },	/* '7' => '`' */
	{ 0x38, 0x5c },	/* '8' => '\\' */
	{ 0x3d, 0x7d },	/* '=' => '}' */
	{ 0x41, 0x40 },	/* 'A' => '@' */
	{ 0x5c, 0x60 },	/* '\\' => '`' */
	{ 0x5e, 0x7c },	/* '^' => '|' */
	{ 0x5f, 0x5d },	/* '_' => ']' */
	{ 0x61, 0x40 },	/* 'a' => '@' */
	{ 0xdc, 0x7c },	/* Pseudo-'\\' => '|' */
	{ 0, 0 }
};

/** "fr" keyboard map */
struct keymap fr_keymap __keymap = {
	.name = "fr",
	.basic = fr_basic,
	.altgr = fr_altgr,
};
