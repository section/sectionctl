# Builds a Windows installer with NSIS.
# It expects the following command line arguments:
# - OUTPUTFILE, filename of the installer (without extension)
# - MAJORVERSION, major build version
# - MINORVERSION, minor build version
# - SECTIONCTL_VERSION, git tag
# - BUILDVERSION, build id version
#
# The created installer executes the following steps:
# 1. install sectionctl for all users
# 2. create an uninstaller
# 3. create geth, attach and uninstall start menu entries
# 4. configures the registry that allows Windows to manage the package through its platform tools
# 5. adds the install directory to %PATH%
#
# Requirements:
# - NSIS, http://nsis.sourceforge.net/Main_Page
# - NSIS Large Strings build, http://nsis.sourceforge.net/Special_Builds
# - SFP, http://nsis.sourceforge.net/NSIS_Simple_Firewall_Plugin (put dll in NSIS\Plugins\x86-ansi)
#
# After intalling NSIS extra the NSIS Large Strings build zip and replace the makensis.exe and the
# files found in Stub.
#
# based on: http://nsis.sourceforge.net/A_simple_installer_with_start_menu_shortcut_and_uninstaller
#
# TODO:
# - sign installer
CRCCheck on

!define GROUPNAME "Section"
!define APPNAME "sectionctl"
!define DESCRIPTION "Section commandline tool"
!addplugindir .\

# Require admin rights on NT6+ (When UAC is turned on)
RequestExecutionLevel admin

# Use LZMA compression
SetCompressor /SOLID lzma

!include LogicLib.nsh

!macro VerifyUserIsAdmin
UserInfo::GetAccountType
pop $0
${If} $0 != "admin" # Require admin rights on NT4+
  messageBox mb_iconstop "Administrator rights required!"
  setErrorLevel 740 # ERROR_ELEVATION_REQUIRED
  quit
${EndIf}
!macroend

function .onInit
  # make vars are global for all users since geth is installed global
  setShellVarContext all
  !insertmacro VerifyUserIsAdmin

  ${If} ${ARCH} == "amd64"
    StrCpy $InstDir "$PROGRAMFILES64\${GROUPNAME}\${APPNAME}"
  ${Else}
    StrCpy $InstDir "$PROGRAMFILES32\${GROUPNAME}\${APPNAME}"
  ${Endif}
functionEnd

!include nsis.install.nsh
!include nsis.uninstall.nsh