export const i18n = {
  appTitle: "Мониторинг системы",
  settings: "Настройки",
  dashboard: "Панель",
  live: "Онлайн",
  offline: "Нет связи",
  lastUpdate: "Обновлено",
  tempUnavailable: "Датчики недоступны (VPS)",
  period: {
    "1h": "Час",
    "6h": "6 часов",
    "1d": "День",
    "1w": "Неделя",
    "2w": "2 недели",
    "1mo": "Месяц",
  },
  status: {
    ok: "Норма",
    warning: "Предупреждение",
    critical: "Критично",
    error: "Ошибка",
    unknown: "Нет данных",
  },
  settingsPage: {
    title: "Настройки мониторинга",
    sensors: "Датчики",
    dashboard: "Панели графиков",
    save: "Сохранить",
    saved: "Настройки сохранены",
    saveError: "Ошибка сохранения",
    enabled: "Включён",
    name: "Название",
    type: "Тип",
    interval: "Интервал (с)",
    warn: "Предупр.",
    critical: "Критич.",
    panelTitle: "Заголовок",
    panelType: "Тип панели",
    sensorsList: "Датчики",
    row: "Строка",
    col: "Колонка",
    span: "Ширина",
    period: "Период",
    addPanel: "Добавить панель",
    addSensor: "Добавить датчик",
    delete: "Удалить",
    noAvailableSensors: "Нет доступных датчиков для этой платформы",
    params: "Параметры",
    unit: "Единица",
    retention: "Хранение истории (дней)",
    defaultInterval: "Интервал по умолчанию (с)",
    availableHint: "Отображаются только датчики, доступные на текущей платформе",
    regenerateToken: "Новый token",
    copyToken: "Копировать",
    tokenCopied: "Token скопирован",
    hubDomain: "Домен hub",
    hubPublicUrl: "Публичный URL",
    hubUrlForAgents: "URL для агентов",
    copyHubUrl: "Копировать URL",
    hubUrlCopied: "URL hub скопирован",
  },
  noData: "Нет данных",
  periodLabel: "Период графиков",
  expand: "Развернуть",
  collapse: "Свернуть",
  spanLabel: { 1: "1 кол.", 2: "2 кол.", 4: "4 кол." },
  systemInfo: {
    title: "Система",
    hostname: "Компьютер",
    os: "ОС",
    cpu: "Процессор",
    cores: "Ядра",
    memory: "Оперативная память",
    swap: "Подкачка",
    disks: "Разделы и тома",
    physicalDrives: "Физические диски",
    partitions: "Разделы",
    gpus: "Видеокарты",
    network: "Сеть",
    battery: "Батарея",
    uptime: "Время работы",
    free: "свободно",
    of: "из",
    usage: "загрузка",
    mediaType: "тип",
    interface: "интерфейс",
    health: "состояние",
    driver: "драйвер",
    cuda: "CUDA",
    cudaToolkit: "toolkit",
    vram: "видеопамять",
    resolution: "разрешение",
    plugged: "от сети",
    onBattery: "от батареи",
    perCore: "по ядрам",
    noGpu: "Видеокарта не обнаружена",
    noNetwork: "Сетевые интерфейсы не найдены",
    noDrives: "Физические диски не обнаружены",
    noPartitions: "Разделы не найдены",
  },
  mediaTypes: {
    HDD: "HDD",
    SSD: "SSD",
    NVMe: "NVMe",
    unknown: "неизвестно",
  },
  footer: {
    loading: "Проверка версии…",
    upToDate: "Актуальная версия установлена",
    updateAvailable: "Доступна новая версия",
    updateCheckFailed: "Не удалось проверить обновления",
  },
  hosts: {
    title: "Хосты",
    name: "Имя",
    hostname: "Hostname",
    status: "Статус",
    lastSeen: "Последний контакт",
    empty: "Агенты ещё не подключались",
    allHosts: "Все хосты",
    selectHost: "Хост",
  },
};

export function formatTime(ts) {
  if (!ts) return "—";
  const date = new Date(ts * 1000);
  return date.toLocaleString("ru-RU");
}

export function formatValue(value, unit) {
  if (value === null || value === undefined) return i18n.noData;
  const formatted = Number.isInteger(value) ? value : Number(value).toFixed(2);
  return unit ? `${formatted} ${unit}` : String(formatted);
}

export function formatUptime(seconds) {
  if (!seconds) return "—";
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  if (days > 0) return `${days} д ${hours} ч`;
  if (hours > 0) return `${hours} ч ${minutes} мин`;
  return `${minutes} мин`;
}
