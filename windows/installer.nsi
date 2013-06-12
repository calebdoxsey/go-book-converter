!include "x64.nsh"
!include "FileAssociation.nsh"
!include "winmessages.nsh"

Name "Go"
OutFile "go-install.exe"
InstallDir "C:\Go"
InstallDirRegKey HKLM "Software\Go" "Install_Dir"
RequestExecutionLevel admin

Page directory
Page instfiles

UninstPage uninstConfirm
UninstPage instfiles

Section
  ; Go
  SetOutPath "$INSTDIR"
  ${If} ${RunningX64}
    File /r "D:\projects\go\src\bitbucket.org\badgerodon\golang-book\installers\windows\go1.0.2.windows-amd64\go\*"
  ${Else}
    File /r "D:\projects\go\src\bitbucket.org\badgerodon\golang-book\installers\windows\go1.0.2.windows-386\go\*"
  ${EndIf}
  
  ; Scite
  SetOutPath "$INSTDIR\scite"
  File /r "D:\projects\go\src\bitbucket.org\badgerodon\golang-book\installers\windows\scite\*"
  ${registerExtension} "$INSTDIR\scite\scite.exe" ".go" "Go File"
  
  ; Home
  SetOutPath "$PROFILE\Go"
  File /r "D:\projects\go\src\bitbucket.org\badgerodon\golang-book\installers\windows\home\*"  

  SetOutPath "$INSTDIR"
  
  ; Environment
  File "D:\projects\go\src\bitbucket.org\badgerodon\golang-book\installers\windows\w32env.exe"
  ExecWait '"$INSTDIR\w32env.exe" add PATH "$INSTDIR\bin"'
  ExecWait '"$INSTDIR\w32env.exe" add PATH "$INSTDIR\scite"'  
  ExecWait '"$INSTDIR\w32env.exe" add PATH "$PROFILE\Go\bin"'
  ExecWait '"$INSTDIR\w32env.exe" set GOROOT "$INSTDIR"'
  ExecWait '"$INSTDIR\w32env.exe" set GOPATH "$PROFILE\Go"'
  
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Go" "DisplayName" "Go"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Go" "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Go" "NoModify" 1
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Go" "NoRepair" 1
  WriteUninstaller "uninstall.exe"
  
  ; Install Font
  File "D:\projects\go\src\bitbucket.org\badgerodon\golang-book\installers\windows\Inconsolata.otf"
  CopyFiles "$INSTDIR\*.otf" "$FONTS"
  WriteRegStr HKLM "Software\Microsoft\Windows NT\CurrentVersion\Fonts" "Inconsolata (TrueType)" "Inconsolata.otf"
  System::Call "GDI32::AddFontResourceA(t) i ('Inconsolata.otf') .s"
  
  ; Shortcuts
  CreateDirectory "$SMPROGRAMS\Go"
  ${If} ${RunningX64}
    CreateShortCut "$SMPROGRAMS\Go\Go Doc Server.lnk" 'C:\Windows\SysWOW64\cmd.exe' '/c start "Godoc Server http://localhost:6060" "$INSTDIR\bin\godoc.exe" -http=localhost:6060 -goroot="$INSTDIR\." && start http://localhost:6060'
  ${Else}
    CreateShortCut "$SMPROGRAMS\Go\Go Doc Server.lnk" 'C:\Windows\System32\cmd.exe' '/c start "Godoc Server http://localhost:6060" "$INSTDIR\bin\godoc.exe" -http=localhost:6060 -goroot="$INSTDIR\." && start http://localhost:6060'
  ${EndIf}
  CreateShortCut "$SMPROGRAMS\Go\Scite.lnk" "$INSTDIR\scite\scite.exe" "" "$INSTDIR\scite\scite.exe" 0
  CreateShortCut "$SMPROGRAMS\Go\Uninstall.lnk" "$INSTDIR\uninstall.exe" "" "$INSTDIR\uninstall.exe" 0
  
  SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
SectionEnd

Section "Uninstall"
  ExecWait '"$INSTDIR\w32env.exe" delete GOPATH'
  ExecWait '"$INSTDIR\w32env.exe" delete GOROOT'
  ExecWait '"$INSTDIR\w32env.exe" remove PATH "$PROFILE\Go\bin"'
  ExecWait '"$INSTDIR\w32env.exe" remove PATH "$INSTDIR\scite"' 
  ExecWait '"$INSTDIR\w32env.exe" remove PATH "$INSTDIR\bin"' 
  SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
  
  ${unregisterExtension} ".go" "Go File"
  ; Remove registry keys
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Go"
  DeleteRegKey HKLM "Software\Go"
  
  ; Remove shortcuts, if any
  Delete "$SMPROGRAMS\Go\*.*"
  
  ; Remove directories used
  RMDir "$SMPROGRAMS\Go"
  RMDir /r "$INSTDIR"
SectionEnd