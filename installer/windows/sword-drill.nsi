!include "MUI2.nsh"

!ifndef VERSION
!define VERSION "dev"
!endif

!ifndef OUTPUT
!define OUTPUT "sword-drill-windows-amd64-setup.exe"
!endif

; Path to the built sword-drill.exe to package. Override with /DEXE_PATH=...
; Default is relative to this script (installer/windows/) so makensis can be
; run directly from the repo root after `wails build`.
!ifndef EXE_PATH
!define EXE_PATH "..\..\build\bin\sword-drill.exe"
!endif

!define APP_NAME "Sword Drill"
!define COMPANY_NAME "Jonathan Beechner"
!define INSTALL_DIR "$PROGRAMFILES64\Sword Drill"
!define UNINSTALL_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"

Name "${APP_NAME}"
OutFile "${OUTPUT}"
InstallDir "${INSTALL_DIR}"
InstallDirRegKey HKLM "${UNINSTALL_KEY}" "InstallLocation"
RequestExecutionLevel admin
Unicode True
SetCompressor /SOLID lzma

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

!insertmacro MUI_LANGUAGE "English"

Section "Install"
    SetOutPath "$INSTDIR"
    SetOverwrite on

    File "/oname=sword-drill.exe" "${EXE_PATH}"
    WriteUninstaller "$INSTDIR\Uninstall.exe"

    CreateDirectory "$SMPROGRAMS\Sword Drill"
    CreateShortcut "$SMPROGRAMS\Sword Drill\Sword Drill.lnk" "$INSTDIR\sword-drill.exe"
    CreateShortcut "$SMPROGRAMS\Sword Drill\Uninstall Sword Drill.lnk" "$INSTDIR\Uninstall.exe"
    CreateShortcut "$DESKTOP\Sword Drill.lnk" "$INSTDIR\sword-drill.exe"

    WriteRegStr HKLM "${UNINSTALL_KEY}" "DisplayName" "${APP_NAME}"
    WriteRegStr HKLM "${UNINSTALL_KEY}" "DisplayVersion" "${VERSION}"
    WriteRegStr HKLM "${UNINSTALL_KEY}" "Publisher" "${COMPANY_NAME}"
    WriteRegStr HKLM "${UNINSTALL_KEY}" "InstallLocation" "$INSTDIR"
    WriteRegStr HKLM "${UNINSTALL_KEY}" "DisplayIcon" "$INSTDIR\sword-drill.exe"
    WriteRegStr HKLM "${UNINSTALL_KEY}" "UninstallString" '"$INSTDIR\Uninstall.exe"'
    WriteRegDWORD HKLM "${UNINSTALL_KEY}" "NoModify" 1
    WriteRegDWORD HKLM "${UNINSTALL_KEY}" "NoRepair" 1
SectionEnd

Section "Uninstall"
    Delete "$DESKTOP\Sword Drill.lnk"
    Delete "$SMPROGRAMS\Sword Drill\Sword Drill.lnk"
    Delete "$SMPROGRAMS\Sword Drill\Uninstall Sword Drill.lnk"
    RMDir "$SMPROGRAMS\Sword Drill"

    Delete "$INSTDIR\sword-drill.exe"
    Delete "$INSTDIR\Uninstall.exe"
    RMDir "$INSTDIR"

    DeleteRegKey HKLM "${UNINSTALL_KEY}"
SectionEnd