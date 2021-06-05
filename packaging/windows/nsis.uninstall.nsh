Section "Uninstall"
  # uninstall for all users
  setShellVarContext all

  Delete $INSTDIR\uninstall.exe

  # Delete install directory
  rmDir $INSTDIR

  # Delete start menu launcher
  Delete "$SMPROGRAMS\${APPNAME}\Uninstall.lnk"
  rmDir "$SMPROGRAMS\${APPNAME}"


  # Remove install directory from PATH
  EnVar::DeleteValue PATH "$INSTDIR"
;  ${un.EnvVarUpdate} $0 "PATH" "A" "HKLM" "$INSTDIR"

  # Cleanup registry (deletes all sub keys)
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${GROUPNAME} ${APPNAME}"
SectionEnd