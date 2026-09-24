; Inno Setup script for system-monitor-agent (Windows)
; Build via packaging/windows/build-windows-agent.ps1

#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif
#ifndef SourceDir
  #define SourceDir "dist\system-monitor-agent"
#endif
#ifndef RepoRoot
  #define RepoRoot "..\.."
#endif
#ifndef OutputDir
  #define OutputDir "output"
#endif

#define MyAppName "system-monitor agent"
#define MyAppPublisher "pagbest154-cmd"
#define MyAppURL "https://github.com/pagbest154-cmd/system-monitor"
#define MyServiceName "system-monitor-agent"
#define MyConfigDir "{commonappdata}\system-monitor"

[Setup]
AppId={{8F4C2A91-5D3E-4B6C-9A1F-2E7D8C4B6F10}
AppName={#MyAppName}
AppVersion={#AppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}/releases
DefaultDirName={autopf}\system-monitor-agent
DefaultGroupName=system-monitor
DisableProgramGroupPage=yes
OutputDir={#OutputDir}
OutputBaseFilename=system-monitor-agent_{#AppVersion}_setup
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=admin
ArchitecturesInstallIn64BitMode=x64compatible
UninstallDisplayIcon={app}\system-monitor-agent.exe
SetupLogging=yes

[Languages]
Name: "russian"; MessagesFile: "compiler:Languages\Russian.isl"
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "trayautostart"; Description: "Запускать иконку в трее при входе в Windows"; GroupDescription: "{cm:AdditionalIcons}"; Flags: checked

[Files]
Source: "{#SourceDir}\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "{#RepoRoot}\config\agent_sensors.yaml"; DestDir: "{#MyConfigDir}"; Flags: onlyifdoesntexist
Source: "third_party\nssm\win64\nssm.exe"; DestDir: "{app}\nssm"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\system-monitor-agent.exe"; Parameters: "--tray"
Name: "{group}\Настройки агента"; Filename: "{app}\system-monitor-agent.exe"; Parameters: "--settings"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\system-monitor-agent.exe"; Parameters: "--tray"; Tasks: desktopicon

[Registry]
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "{#MyServiceName}-tray"; ValueData: """{app}\system-monitor-agent.exe"" --tray"; Flags: uninsdeletevalue; Tasks: trayautostart

[UninstallRun]
Filename: "{app}\nssm\nssm.exe"; Parameters: "stop {#MyServiceName}"; Flags: runhidden; RunOnceId: "StopService"
Filename: "{app}\nssm\nssm.exe"; Parameters: "remove {#MyServiceName} confirm"; Flags: runhidden; RunOnceId: "RemoveService"

[Code]
var
  HubUrlPage: TInputQueryWizardPage;
  AgentIdPage: TInputQueryWizardPage;
  TokenPage: TInputQueryWizardPage;

function ConfigFilePath: String;
begin
  Result := ExpandConstant('{#MyConfigDir}\agent.yaml');
end;

function TokenFilePath: String;
begin
  Result := ExpandConstant('{#MyConfigDir}\agent.token');
end;

function ConfigExists: Boolean;
begin
  Result := FileExists(ConfigFilePath);
end;

function EscapeYaml(Value: String): String;
begin
  StringChangeEx(Value, '\', '\\', True);
  StringChangeEx(Value, '"', '\"', True);
  Result := Value;
end;

procedure InitializeWizard;
begin
  HubUrlPage := CreateInputQueryPage(wpWelcome,
    'Подключение к Hub', 'Укажите адрес центрального hub',
    'Введите URL hub, например http://192.168.1.10:8080');
  HubUrlPage.Add('Hub URL:', False);
  HubUrlPage.Values[0] := 'http://127.0.0.1:8080';

  AgentIdPage := CreateInputQueryPage(HubUrlPage.ID,
    'Идентификатор агента', 'Уникальный Agent ID',
    'По умолчанию используется имя компьютера.');
  AgentIdPage.Add('Agent ID:', False);
  AgentIdPage.Values[0] := '';

  TokenPage := CreateInputQueryPage(AgentIdPage.ID,
    'Токен агента', 'Токен из config/agents.yaml на hub',
    'Оставьте пустым, если токен будет добавлен позже.');
  TokenPage.Add('Token (optional):', True);
  TokenPage.Values[0] := '';
end;

function ShouldSkipPage(PageID: Integer): Boolean;
begin
  if ConfigExists then
  begin
    Result := (PageID = HubUrlPage.ID) or (PageID = AgentIdPage.ID) or (PageID = TokenPage.ID);
    Exit;
  end;
  Result := False;
end;

function WriteAgentConfig: Boolean;
var
  HubUrl, AgentId, Token, TokenFile, ConfigFile: String;
  Lines: TArrayOfString;
begin
  if ConfigExists then
  begin
    Result := True;
    Exit;
  end;

  HubUrl := Trim(HubUrlPage.Values[0]);
  if HubUrl = '' then
    HubUrl := 'http://127.0.0.1:8080';

  AgentId := Trim(AgentIdPage.Values[0]);
  if AgentId = '' then
    AgentId := ExpandConstant('{computername}');

  Token := TokenPage.Values[0];
  TokenFile := TokenFilePath;
  ConfigFile := ConfigFilePath;

  ForceDirectories(ExpandConstant('{#MyConfigDir}'));

  SetArrayLength(Lines, 5);
  Lines[0] := 'hub_url: "' + EscapeYaml(HubUrl) + '"';
  Lines[1] := 'agent_id: "' + EscapeYaml(AgentId) + '"';
  Lines[2] := 'token_file: ' + TokenFile;
  Lines[3] := 'interval_sec: 5';
  Lines[4] := 'transport: http';

  if not SaveStringsToFile(ConfigFile, Lines, False) then
  begin
    MsgBox('Не удалось записать agent.yaml', mbError, MB_OK);
    Result := False;
    Exit;
  end;

  if Token <> '' then
  begin
    if not SaveStringToFile(TokenFile, Token, False) then
    begin
      MsgBox('Не удалось записать agent.token', mbError, MB_OK);
      Result := False;
      Exit;
    end;
  end
  else if not FileExists(TokenFile) then
  begin
    SaveStringToFile(TokenFile, '', False);
  end;

  Result := True;
end;

function InstallWindowsService: Boolean;
var
  ResultCode: Integer;
  Nssm, Exe, Config, AppDir, LogDir: String;
begin
  Nssm := ExpandConstant('{app}\nssm\nssm.exe');
  Exe := ExpandConstant('{app}\system-monitor-agent.exe');
  Config := ConfigFilePath;
  AppDir := ExpandConstant('{app}');
  LogDir := ExpandConstant('{#MyConfigDir}');

  Exec(Nssm, 'stop {#MyServiceName}', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Exec(Nssm, 'remove {#MyServiceName} confirm', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);

  if not Exec(Nssm,
    'install {#MyServiceName} "' + Exe + '" --config "' + Config + '"',
    '', SW_HIDE, ewWaitUntilTerminated, ResultCode) or (ResultCode <> 0) then
  begin
    MsgBox('Не удалось установить службу Windows (код ' + IntToStr(ResultCode) + ').', mbError, MB_OK);
    Result := False;
    Exit;
  end;

  Exec(Nssm, 'set {#MyServiceName} AppDirectory "' + AppDir + '"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Exec(Nssm, 'set {#MyServiceName} DisplayName "{#MyAppName}"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Exec(Nssm, 'set {#MyServiceName} Description "Metrics collector for system-monitor hub"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Exec(Nssm, 'set {#MyServiceName} Start SERVICE_AUTO_START', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Exec(Nssm, 'set {#MyServiceName} AppStdout "' + LogDir + '\agent.log"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Exec(Nssm, 'set {#MyServiceName} AppStderr "' + LogDir + '\agent.err.log"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Exec(Nssm, 'set {#MyServiceName} AppRotateFiles 1', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Exec(Nssm, 'set {#MyServiceName} AppRotateBytes 1048576', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);

  if not Exec(Nssm, 'start {#MyServiceName}', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) or (ResultCode <> 0) then
  begin
    MsgBox('Служба установлена, но не запустилась (код ' + IntToStr(ResultCode) + '). Проверьте agent.yaml и логи.', mbInformation, MB_OK);
  end;

  Result := True;
end;

procedure LaunchTrayIcon;
var
  ResultCode: Integer;
begin
  Exec(ExpandConstant('{app}\system-monitor-agent.exe'), '--tray', '', SW_SHOWMINNOACTIVE, ewNoWait, ResultCode);
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    if not WriteAgentConfig then
      Abort;
    if not InstallWindowsService then
      Abort;
    LaunchTrayIcon;
  end;
end;

procedure CurUninstallStep(UninstallStep: TUninstallStep);
var
  ResultCode: Integer;
  Nssm: String;
begin
  if UninstallStep = usUninstall then
  begin
    Nssm := ExpandConstant('{app}\nssm\nssm.exe');
    if FileExists(Nssm) then
    begin
      Exec(Nssm, 'stop {#MyServiceName}', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
      Exec(Nssm, 'remove {#MyServiceName} confirm', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    end;
  end;
end;
