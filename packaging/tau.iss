; The Windows installer.
;
; It carries the same tree every other package carries: the interpreter in
; bin, the standard library and the shared objects it opens in lib\tau. The
; interpreter finds that library by looking beside itself, so nothing here
; sets TAUPATH and an install moved somewhere else still works.
;
; Built with:
;	iscc /DTauVersion=2.1.0 /DStageDir=dist\win packaging\tau.iss

#ifndef TauVersion
  #define TauVersion "0.0.0"
#endif
#ifndef StageDir
  #define StageDir "..\dist\win"
#endif

[Setup]
AppId={{4F0B2C3E-9C2A-4E31-8B77-7A2D0B5F9E10}
AppName=Tau
AppVersion={#TauVersion}
AppPublisher=Nicolo Santamaria
AppPublisherURL=https://github.com/NicoNex/tau
AppSupportURL=https://github.com/NicoNex/tau/issues
DefaultDirName={autopf}\Tau
DefaultGroupName=Tau
DisableProgramGroupPage=yes
LicenseFile={#StageDir}\LICENSE
OutputBaseFilename=tau-setup-windows-x86_64
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
; A per user install needs no administrator, and asking for one is the
; difference between an installer that runs and one that is shrugged at.
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog commandline
UninstallDisplayName=Tau {#TauVersion}

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "addtopath"; Description: "Add Tau to the PATH, so that 'tau' works in any terminal"; GroupDescription: "Setting up:"

[Files]
Source: "{#StageDir}\bin\tau.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
; The library, everything under it: the tau sources and the DLLs half of them
; open. recursesubdirs is what carries sync\atomic and net\http, and a package
; without them is an interpreter that fails on the first import.
Source: "{#StageDir}\lib\tau\*"; DestDir: "{app}\lib\tau"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "{#StageDir}\LICENSE"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#StageDir}\README.md"; DestDir: "{app}"; Flags: ignoreversion isreadme

[Icons]
Name: "{group}\Tau"; Filename: "{app}\bin\tau.exe"
Name: "{group}\Uninstall Tau"; Filename: "{uninstallexe}"

[Registry]
; The PATH of whoever installed it, or of the machine when it was installed
; for everyone.
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; \
    ValueData: "{olddata};{app}\bin"; Check: NeedsPath('{app}\bin') and not IsAdminInstallMode; Tasks: addtopath
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; \
    ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}\bin"; \
    Check: NeedsPath('{app}\bin') and IsAdminInstallMode; Tasks: addtopath

[Run]
Filename: "{app}\bin\tau.exe"; Parameters: "version"; Description: "Show the version"; \
    Flags: postinstall nowait skipifsilent runasoriginaluser

[Code]
{ Whether the directory is already on the PATH, so that installing twice does
  not write it twice. }
function NeedsPath(Param: string): boolean;
var
  Path: string;
  Root: integer;
  Key: string;
begin
  if IsAdminInstallMode then
  begin
    Root := HKEY_LOCAL_MACHINE;
    Key := 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment';
  end
  else
  begin
    Root := HKEY_CURRENT_USER;
    Key := 'Environment';
  end;

  if not RegQueryStringValue(Root, Key, 'Path', Path) then
  begin
    Result := True;
    exit;
  end;

  Result := Pos(';' + Uppercase(ExpandConstant(Param)) + ';', ';' + Uppercase(Path) + ';') = 0;
end;
