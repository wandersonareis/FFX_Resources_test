package util

const FFX_CODE_TABLE_NAME = "ffx.tbs"
const FFX2_CODE_TABLE_NAME = "ffx2.tbs"
const CHARACTER_CODE_TABLE = "ffx_chars.tbs"
const CHARACTER_LOC_CODE_TABLE = "ffxloc.tbs"
const DIALOG_HANDLER_APPLICATION = "ffxdlg_new.exe"
const DIALOG_SPECIAL_HANDLER_APPLICATION = "ffxdlg_ignore_F0FF_size_flags.exe"
const FFX_KERNEL_HANDLER_APPLICATION = "ffxmt.exe"
const FFX2_KERNEL_HANDLER_APPLICATION = "ffx2mt.exe"
const DCP_FILE_XPLITTER_APPLICATION = "SHSplit.exe"
const LOCKIT_HANDLER_APPLICATION = "fcopy.exe"

const UTF8BOM_NORMALIZER_APPLICATION = "utf8bom.exe"

const DEFAULT_APPLICATION_EXTENSION = ".exe"
const DEFAULT_TABLE_EXTENSION = ".tbs"
const DEFAULT_RESOURCES_ROOTDIR = "bin"

const DIALOGS_TARGET_DIR_NAME = "dialogs_text"
const KERNEL_TARGET_DIR_NAME = "kernel_text"

const DCP_TXT_PARTS_PATTERN = dcp_file_pattern + "\\.txt"
const DCP_FILE_PARTS_PATTERN = dcp_file_pattern + "$"
const DCP_PARTS_TARGET_DIR_NAME = "system_text"
const dcp_file_pattern = "macrodic.*?\\.00[0-6]"

const LOCKIT_NAME_BASE = "loc_kit_ps3"
const LOCKIT_TARGET_DIR_NAME = "lockit_text"
const LOCKIT_FILE_PARTS_PATTERN = `.*loc_kit_ps3.*\.part([0-9]{2})$`
const LOCKIT_TXT_PARTS_PATTERN = `.*loc_kit_ps3.*\.part([0-9]{2}).*\.txt$`
