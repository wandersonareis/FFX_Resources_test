package util

const CHARACTER_TABLE_RESOURCES_DIR string = "tbs"
const FFX_ENCODING_TABLE_NAME string = "ffx.tbs"
const FFX2_ENCODING_TABLE_NAME string = "ffx2.tbs"
const CHARACTER_ENCODING_TABLE string = "ffx_chars.tbs"
const CHARACTER_LOC_ENCODING_TABLE string = "ffxloc.tbs"

const DIALOG_HANDLER_RESOURCES_DIR string = "dlg"
const DIALOG_HANDLER_APPLICATION string = "ffxdlg_new.exe"
const DIALOG_SPECIAL_HANDLER_APPLICATION string = "ffxdlg_ignore_F0FF_size_flags.exe"

const KERNEL_HANDLER_RESOURCES_DIR string = "mt"
const FFX_KERNEL_HANDLER_APPLICATION string = "ffxmt.exe"
const FFX2_KERNEL_HANDLER_APPLICATION string = "ffx2mt.exe"

const UTILS_RESOURCES_DIR string = "utils"
const DCP_FILE_XPLITTER_APPLICATION string = "SHSplit.exe"
const LOCKIT_HANDLER_APPLICATION string = "fcopy.exe"
const UTF8BOM_NORMALIZER_APPLICATION string = "utf8bom.exe"

const DEFAULT_APPLICATION_FILE_EXTENSION string = ".exe"
const DEFAULT_ENCODING_TABLE_FILE_EXTENSION string = ".tbs"

//const DEFAULT_RESOURCES_ROOTDIR = "bin"

const DIALOGS_TARGET_DIR_NAME string = "dialogs_text"
const KERNEL_TARGET_DIR_NAME string = "kernel_text"

const DCP_TXT_PARTS_PATTERN string = dcp_file_pattern + "\\.txt"
const DCP_FILE_PARTS_PATTERN string = dcp_file_pattern + "$"
const DCP_PARTS_TARGET_DIR_NAME string = "system_text"
const dcp_file_pattern string = "macrodic.*?\\.00[0-6]"

const LOCKIT_NAME_BASE string = "loc_kit_ps3"
const LOCKIT_TARGET_DIR_NAME string = "lockit_text"
const LOCKIT_FILE_PARTS_PATTERN string = `.*loc_kit_ps3.*\.part([0-9]{2})$`
const LOCKIT_TXT_PARTS_PATTERN string = `.*loc_kit_ps3.*\.part([0-9]{2}).*\.txt$`
