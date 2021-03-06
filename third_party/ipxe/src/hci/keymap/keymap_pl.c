/** @file
 *
 * "pl" keyboard mapping
 *
 * This file is automatically generated; do not edit
 *
 */

FILE_LICENCE ( PUBLIC_DOMAIN );

#include <ipxe/keymap.h>

/** "pl" basic remapping */
static struct keymap_key pl_basic[] = {
	{ 0xdc, 0x3c },	/* Pseudo-'\\' => '<' */
	{ 0xfc, 0x3e },	/* Pseudo-'|' => '>' */
	{ 0, 0 }
};

/** "pl" AltGr remapping */
static struct keymap_key pl_altgr[] = {
	{ 0, 0 }
};

/** "pl" keyboard map */
struct keymap pl_keymap __keymap = {
	.name = "pl",
	.basic = pl_basic,
	.altgr = pl_altgr,
};
