package common

const FFX_DIR_MARKER = "ffx_ps2"
const FFX_TEXT_FORMAT_SEPARATOR = "--------------------------------------------------End--"
const DIALOGS_TARGET_DIR_NAME = "dialogs_text"
const KERNEL_TARGET_DIR_NAME = "kernel_text"
const DCP_PARTS_TARGET_DIR_NAME = "system_text"
const LOCKIT_TARGET_DIR_NAME = "lockit_text"

const FFX_CODE_TABLE_NAME = "ffx.tbs"
const FFX2_CODE_TABLE_NAME = "ffx2.tbs"
const CHARACTER_CODE_TABLE = "ffx_chars.tbs"
const CHARACTER_LOC_CODE_TABLE = "ffxloc.tbs"
const DIALOG_HANDLER_APPLICATION = "ffxdlg_new.exe"
const KERNEL_HANDLER_APPLICATION = "ffx2mt.exe"
const DCP_FILE_XPLITTER_APPLICATION = "SHSplit.exe"
const LOCKIT_HANDLER_APPLICATION = "fcopy.exe"

const UTF8BOM_NORMALIZER_APPLICATION = "utf8bom.exe"

const DEFAULT_APPLICATION_EXTENSION = ".exe"
const DEFAULT_RESOURCES_ROOTDIR = "bin"

const MACRODIC_PATTERN = "macrodic.*?\\.00[0-6]"

const LOCKIT_NAME_BASE = "loc_kit_ps3"
const LOCKIT_FILE_PARTS_PATTERN = `.*loc_kit_ps3.*\.part([0-9]{2})$`
const LOCKIT_TXT_PARTS_PATTERN = `.*loc_kit_ps3.*\.part([0-9]{2}).*\.txt$`
