<script setup>import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import FileUploader from '@/components/FileUploader.vue';
import { useLogsStore } from '@/stores/logs';
const logsStore = useLogsStore();
const sourceText = ref('');
const targetText = ref('');
const sourceLang = ref('zh');
const targetLang = ref('en');
const isTranslating = ref(false);
const importResult = ref([]);
const showImportPanel = ref(false);
const languageOptions = [
 { code: 'zh', name: '中文', native: '中文' },
 { code: 'en', name: 'English', native: 'English' },
 { code: 'ar', name: 'Arabic', native: 'العربية' },
 { code: 'ru', name: 'Russian', native: 'Русский' },
 { code: 'ja', name: 'Japanese', native: '日本語' },
 { code: 'ko', name: 'Korean', native: '한국어' },
 { code: 'fr', name: 'French', native: 'Français' },
 { code: 'de', name: 'German', native: 'Deutsch' }
];
const glossary = {
 '起爆器': { en: 'Initiator', ar: 'مبدد', ru: 'Инициатор', ja: '起爆器', ko: '시동기', fr: 'Initiateur', de: 'Initiator' },
 '导爆管': { en: 'Detonating Cord', ar: 'حبل التفجير', ru: 'Дetonating Cord', ja: '導爆管', ko: '도폭관', fr: 'Cordon détonant', de: 'Detonationsschnur' },
 '延时时间': { en: 'Delay Time', ar: 'وقت التأخير', ru: 'Время задержки', ja: '遅延時間', ko: '지연 시간', fr: 'Temps de retard', de: 'Verzögerungszeit' },
 '起爆网络': { en: 'Blasting Network', ar: 'شبكة التفجير', ru: 'Сеть взрыва', ja: '起爆ネットワーク', ko: '기폭 네트워크', fr: 'Réseau de détonation', de: 'Sprengnetz' },
 '雷管': { en: 'Detonator', ar: 'مفجر', ru: 'Детонатор', ja: '雷管', ko: '뢰관', fr: 'Détonateur', de: 'Detonator' },
 '炸药': { en: 'Explosive', ar: 'متفجر', ru: 'Взрывчатка', ja: '爆薬', ko: '폭약', fr: 'Explosif', de: 'Sprengstoff' },
 '爆破': { en: 'Blasting', ar: 'تفجير', ru: 'Взрыв', ja: '爆破', ko: '발파', fr: 'Démolition', de: 'Sprengung' },
 '安全距离': { en: 'Safety Distance', ar: 'المسافة الآمنة', ru: 'Безопасное расстояние', ja: '安全距離', ko: '안전 거리', fr: 'Distance de sécurité', de: 'Sicherheitsabstand' },
 '爆破参数': { en: 'Blasting Parameters', ar: 'معلمات التفجير', ru: 'Параметры взрыва', ja: '爆破パラメータ', ko: '발파 매개변수', fr: 'Paramètres de détonation', de: 'Sprengparameter' },
 '通信模块': { en: 'Communication Module', ar: 'وحدة الاتصال', ru: 'Модуль связи', ja: '通信モジュール', ko: '통신 모듈', fr: 'Module de communication', de: 'Kommunikationsmodul' },
 '故障诊断': { en: 'Fault Diagnosis', ar: 'تشخيص العيوب', ru: 'Диагностика неисправностей', ja: '故障診断', ko: '고장 진단', fr: 'Diagnostic de panne', de: 'Fehlerdiagnose' },
 '数据加密': { en: 'Data Encryption', ar: 'تشفير البيانات', ru: 'Шифрование данных', ja: 'データ暗号化', ko: '데이터 암호화', fr: 'Chiffrement des données', de: 'Datenverschlüsselung' },
 '远程控制': { en: 'Remote Control', ar: 'التحكم عن بُعد', ru: 'Удаленное управление', ja: '遠隔操作', ko: '원격 제어', fr: 'Commande à distance', de: 'Fernsteuerung' },
 'AI分析': { en: 'AI Analysis', ar: 'تحليل الذكاء الاصطناعي', ru: 'AI-анализ', ja: 'AI分析', ko: 'AI 분석', fr: 'Analyse AI', de: 'KI-Analyse' },
 '授权码': { en: 'Authorization Code', ar: 'كود التأييد', ru: 'Код авторизации', ja: '認証コード', ko: '인증 코드', fr: 'Code d\'autorisation', de: 'Autorisierungscode' },
 '设备码': { en: 'Device Code', ar: 'كود الجهاز', ru: 'Код устройства', ja: 'デバイスコード', ko: '기기 코드', fr: 'Code de dispositif', de: 'Gerätecode' },
 '密码不能为空': { en: 'Password cannot be empty', ar: 'كلمة المرور لا يمكن أن تكون فارغة', ru: 'Пароль не может быть пустым', ja: 'パスワードは空にできません', ko: '비밀번호는 비워둘 수 없습니다', fr: 'Le mot de passe ne peut pas être vide', de: 'Passwort darf nicht leer sein' },
 '用户名不能为空': { en: 'Username cannot be empty', ar: 'اسم المستخدم لا يمكن أن يكون فارغاً', ru: 'Имя пользователя не может быть пустым', ja: 'ユーザー名は空にできません', ko: '사용자 이름은 비워둘 수 없습니다', fr: 'Le nom d\'utilisateur ne peut pas être vide', de: 'Benutzername darf nicht leer sein' },
 '授权码验证': { en: 'Authorization code verification', ar: 'التحقق من كود التأييد', ru: 'Проверка кода авторизации', ja: '認証コードの検証', ko: '인증 코드 확인', fr: 'Vérification du code d\'autorisation', de: 'Autorisierungscode-Verifizierung' },
 '安全验证': { en: 'Security verification', ar: 'التحقق الأمني', ru: 'Безопасностная проверка', ja: 'セキュリティ検証', ko: '보안 확인', fr: 'Vérification de sécurité', de: 'Sicherheitsverifizierung' },
 '控制器': { en: 'Controller', ar: 'منظم', ru: 'Контроллер', ja: 'コントローラー', ko: '컨트롤러', fr: 'Contrôleur', de: 'Controller' },
 '总线电压': { en: 'Bus Voltage', ar: 'جهد الحافلة', ru: 'Напряжение шины', ja: 'バス電圧', ko: '버스 전압', fr: 'Tension de bus', de: 'Busspannung' },
 '总线电流': { en: 'Bus Current', ar: 'تيار الحافلة', ru: 'Ток шины', ja: 'バス電流', ko: '버스 전류', fr: 'Courant de bus', de: 'Busstrom' },
 '单发检测': { en: 'Single-shot Detection', ar: 'كشف لقطة واحدة', ru: 'Одиночное детектирование', ja: '単発検出', ko: '단발 검출', fr: 'Détection à tir unique', de: 'Einzel-Schuss-Erkennung' },
 '连发检测': { en: 'Multi-shot Detection', ar: 'كشف لقطات متعددة', ru: 'Многоразовое детектирование', ja: '連発検出', ko: '연발 검출', fr: 'Détection à tirs multiples', de: 'Mehrfach-Schuss-Erkennung' },
 '一键翻译': { en: 'One-click Translation', ar: 'ترجمة بنقرة واحدة', ru: 'Одношаговый перевод', ja: 'ワンクリック翻訳', ko: '원클릭 번역', fr: 'Traduction en un clic', de: 'Ein-Klick-Übersetzung' },
 '导出结果': { en: 'Export Results', ar: 'تصدير النتائج', ru: 'Экспорт результатов', ja: '結果をエクスポート', ko: '결과 내보내기', fr: 'Exporter les résultats', de: 'Ergebnisse exportieren' },
 '上传文件': { en: 'Upload File', ar: 'رفع ملف', ru: 'Загрузить файл', ja: 'ファイルをアップロード', ko: '파일 업로드', fr: 'Télécharger un fichier', de: 'Datei hochladen' },
 '返回上一页': { en: 'Back to Previous Page', ar: 'العودة إلى الصفحة السابقة', ru: 'Вернуться на предыдущую страницу', ja: '前のページに戻る', ko: '이전 페이지로 돌아가기', fr: 'Retour à la page précédente', de: 'Zurück zur vorherigen Seite' },
 '网络授时': { en: 'Network Time Sync', ar: 'مزامنة وقت الشبكة', ru: 'Синхронизация сетевого времени', ja: 'ネットワーク時刻同期', ko: '네트워크 시각 동기화', fr: 'Synchronisation temporelle réseau', de: 'Netzwerk-Zeitsynchronisation' },
 '低压检测': { en: 'Low Voltage Detection', ar: 'كشف الجهد المنخفض', ru: 'Обнаружение низкого напряжения', ja: '低電圧検出', ko: '저전압 검출', fr: 'Détection de basse tension', de: 'Niederspannungsdetektion' },
 '实时检测': { en: 'Real-time Detection', ar: 'الكشف في الوقت الفعلي', ru: 'Реальное время детектирования', ja: 'リアルタイム検出', ko: '실시간 검출', fr: 'Détection en temps réel', de: 'Echtzeit-Erkennung' },
 '起爆指令': { en: 'Firing Command', ar: 'أمر الإطلاق', ru: 'Команда на выстрел', ja: '起爆指令', ko: '기폭 명령', fr: 'Commande de tir', de: 'Feuerbefehl' },
 '检测类型': { en: 'Detection Type', ar: 'نوع الكشف', ru: 'Тип детектирования', ja: '検出タイプ', ko: '검출 유형', fr: 'Type de détection', de: 'Erkennungstyp' },
 '状态': { en: 'Status', ar: 'الحالة', ru: 'Статус', ja: 'ステータス', ko: '상태', fr: 'État', de: 'Status' },
 'UID': { en: 'UID', ar: 'UID', ru: 'UID', ja: 'UID', ko: 'UID', fr: 'UID', de: 'UID' },
 '管壳码': { en: 'Shell Code', ar: 'كود الغلاف', ru: 'Код оболочки', ja: 'シェルコード', ko: '셸 코드', fr: 'Code de coque', de: 'Shell-Code' },
 '电流': { en: 'Current', ar: 'تيار', ru: 'Ток', ja: '電流', ko: '전류', fr: 'Courant', de: 'Strom' },
 '电压': { en: 'Voltage', ar: 'جهد', ru: 'Напряжение', ja: '電圧', ko: '전압', fr: 'Tension', de: 'Spannung' }
};
const handleTranslate = async () => {
 if (!sourceText.value.trim()) {
 ElMessage.warning('请输入原文内容');
 return;
 }
 isTranslating.value = true;
 await new Promise(resolve => setTimeout(resolve, 800));
 let result = sourceText.value;
 Object.keys(glossary).forEach(key => {
 const regex = new RegExp(key, 'g');
 if (sourceLang.value === 'zh') {
 const targetTranslation = glossary[key][targetLang.value];
 if (targetTranslation) {
 result = result.replace(regex, targetTranslation);
 }
 } else if (targetLang.value === 'zh') {
 const sourceTranslation = glossary[key][sourceLang.value];
 if (sourceTranslation) {
 const valueRegex = new RegExp(sourceTranslation, 'gi');
 result = result.replace(valueRegex, key);
 }
 }
 });
 const langNames = languageOptions.find(l => l.code === targetLang.value);
 result = `${langNames?.native || targetLang.value}: ${result}`;
 targetText.value = result;
 isTranslating.value = false;
 logsStore.addLog('翻译', 'AI翻译', `方向: ${languageOptions.find(l => l.code === sourceLang.value)?.name} -> ${languageOptions.find(l => l.code === targetLang.value)?.name}`);
 ElMessage.success('翻译完成');
};
const handleSwap = () => {
 const tempLang = sourceLang.value;
 sourceLang.value = targetLang.value;
 targetLang.value = tempLang;
 const temp = sourceText.value;
 sourceText.value = targetText.value;
 targetText.value = temp;
};
const handleCopy = () => {
 if (!targetText.value) {
 ElMessage.warning('没有可复制的内容');
 return;
 }
 navigator.clipboard.writeText(targetText.value).then(() => {
 ElMessage.success('翻译结果已复制');
 logsStore.addLog('复制', 'AI翻译', '复制翻译结果');
 });
};
const handleClear = () => {
 sourceText.value = '';
 targetText.value = '';
 importResult.value = [];
 showImportPanel.value = false;
 logsStore.addLog('清空', 'AI翻译', '清空内容');
};
const parseXML = (content) => {
 const regex = /<string name="([^"]+)">([^<]+)<\/string>/g;
 const results = [];
 let match;
 while ((match = regex.exec(content)) !== null) {
 results.push({
 key: match[1],
 chinese: match[2].trim(),
 translation: ''
 });
 }
 return results;
};
const parseExcelTSV = (content) => {
 const lines = content.split('\n').filter(line => line.trim());
 const results = [];
 for (let i = 1; i < lines.length; i++) {
 const parts = lines[i].split('\t');
 if (parts.length >= 2) {
 results.push({
 key: parts[0]?.trim() || '',
 chinese: parts[1]?.trim() || '',
 translation: parts[2]?.trim() || ''
 });
 }
 }
 return results;
};
const handleFilesSelected = (files) => {
 const file = files[0];
 if (!file)
 return;
 const reader = new FileReader();
 reader.onload = (e) => {
 const content = e.target.result;
 try {
 if (file.name.endsWith('.xml')) {
 importResult.value = parseXML(content);
 }
 else if (file.name.endsWith('.txt') || file.name.endsWith('.tsv')) {
 importResult.value = parseExcelTSV(content);
 }
 else {
 ElMessage.warning('不支持的文件格式，请上传 XML 或 TXT/TSV 文件');
 return;
 }
 showImportPanel.value = true;
 ElMessage.success(`成功导入 ${importResult.value.length} 条记录`);
 logsStore.addLog('导入', 'AI翻译', `导入文件: ${file.name}，共 ${importResult.value.length} 条`);
 }
 catch (e) {
 ElMessage.error('文件解析失败');
 }
 };
 reader.readAsText(file, 'UTF-8');
};
const translateImportedItems = async () => {
 if (importResult.value.length === 0)
 return;
 isTranslating.value = true;
 for (let i = 0; i < importResult.value.length; i++) {
 const item = importResult.value[i];
 let translation = item.chinese;
 Object.keys(glossary).forEach(key => {
 const regex = new RegExp(key, 'g');
 const targetTranslation = glossary[key][targetLang.value];
 if (targetTranslation) {
 translation = translation.replace(regex, targetTranslation);
 }
 });
 item.translation = translation;
 }
 isTranslating.value = false;
 ElMessage.success('批量翻译完成');
 logsStore.addLog('翻译', 'AI翻译', `批量翻译 ${importResult.value.length} 条记录`);
};
const exportXML = () => {
 if (importResult.value.length === 0) {
 ElMessage.warning('没有可导出的数据');
 return;
 }
 let content = '<?xml version="1.0" encoding="UTF-8"?>\n<resources>\n';
 importResult.value.forEach(item => {
 content += ` <string name="${item.key}">${item.translation || item.chinese}</string>\n`;
 });
 content += '</resources>';
 downloadFile(content, 'translations.xml', 'text/xml');
 logsStore.addLog('导出', 'AI翻译', '导出XML文件');
};
const exportTSV = () => {
 if (importResult.value.length === 0) {
 ElMessage.warning('没有可导出的数据');
 return;
 }
 let content = 'KEY\t中文内容\t目标语言\n';
 importResult.value.forEach(item => {
 content += `${item.key}\t${item.chinese}\t${item.translation || ''}\n`;
 });
 downloadFile(content, 'translations.tsv', 'text/tab-separated-values');
 logsStore.addLog('导出', 'AI翻译', '导出TSV文件');
};
const downloadFile = (content, filename, type) => {
 const blob = new Blob(['\uFEFF' + content], { type: `${type};charset=utf-8;` });
 const url = URL.createObjectURL(blob);
 const a = document.createElement('a');
 a.href = url;
 a.download = filename;
 a.click();
 URL.revokeObjectURL(url);
 ElMessage.success(`导出成功: ${filename}`);
};
const getGlossaryList = () => {
 const list = [];
 Object.keys(glossary).forEach(key => {
 list.push({ zh: key, en: glossary[key].en });
 });
 return list;
};
</script>

<template>
  <div class="ai-translate">
    <GlassCard title="AI翻译工具">
      <div class="language-selector">
        <select v-model="sourceLang" class="lang-select">
          <option v-for="lang in languageOptions" :key="lang.code" :value="lang.code">
            {{ lang.name }}
          </option>
        </select>
        <button class="swap-btn-small" @click="handleSwap">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="16 3 21 3 21 8" />
            <line x1="3" y1="12" x2="21" y2="12" />
            <polyline points="8 21 3 21 3 16" />
            <line x1="16" y1="16" x2="21" y2="21" />
            <line x1="8" y1="8" x2="3" y2="3" />
          </svg>
        </button>
        <select v-model="targetLang" class="lang-select">
          <option v-for="lang in languageOptions" :key="lang.code" :value="lang.code">
            {{ lang.name }}
          </option>
        </select>
      </div>
      
      <div class="translate-section">
        <div class="input-panel">
          <div class="panel-header">
            <span class="panel-title">{{ languageOptions.find(l => l.code === sourceLang)?.native || sourceLang }}</span>
          </div>
          <textarea 
            v-model="sourceText"
            class="text-input"
            :placeholder="`请输入${languageOptions.find(l => l.code === sourceLang)?.name}文本...`"
            rows="8"
          ></textarea>
        </div>
        
        <div class="swap-btn-wrap">
          <button class="swap-btn" @click="handleSwap">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="16 3 21 3 21 8" />
              <line x1="3" y1="12" x2="21" y2="12" />
              <polyline points="8 21 3 21 3 16" />
              <line x1="16" y1="16" x2="21" y2="21" />
              <line x1="8" y1="8" x2="3" y2="3" />
            </svg>
          </button>
        </div>
        
        <div class="output-panel">
          <div class="panel-header">
            <span class="panel-title">{{ languageOptions.find(l => l.code === targetLang)?.native || targetLang }}</span>
            <button class="copy-btn" @click="handleCopy">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
              </svg>
              <span>复制</span>
            </button>
          </div>
          <textarea 
            v-model="targetText"
            class="text-input output"
            :placeholder="`翻译结果将在此显示...`"
            rows="8"
            readonly
          ></textarea>
        </div>
      </div>
      
      <div class="action-buttons">
        <button 
          class="translate-btn" 
          :disabled="isTranslating"
          @click="handleTranslate"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="21" y1="1" x2="1" y2="21" />
            <polygon points="16 1 21 1 21 6 16 6 16 1" />
            <polygon points="5 21 5 16 1 16 1 21 5 21" />
          </svg>
          <span>{{ isTranslating ? 'AI翻译中...' : '一键AI翻译' }}</span>
        </button>
        <button class="clear-btn" @click="handleClear">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 6h18" />
            <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
            <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
          </svg>
          <span>清空内容</span>
        </button>
      </div>

      <div class="file-import-section">
        <div class="section-title">批量文件翻译</div>
        <FileUploader @files-selected="handleFilesSelected" />
        <div class="import-hint">支持 XML 格式 (`&lt;string name="key"&gt;中文内容&lt;/string&gt;`) 或 TSV 格式（KEY、中文内容、目标语言）</div>
      </div>

      <div v-if="showImportPanel" class="import-panel">
        <div class="panel-header">
          <span class="panel-title">导入的翻译列表 ({{ importResult.length }} 条)</span>
          <div class="panel-actions">
            <button class="action-btn" @click="translateImportedItems">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="21" y1="1" x2="1" y2="21" />
                <polygon points="16 1 21 1 21 6 16 6 16 1" />
                <polygon points="5 21 5 16 1 16 1 21 5 21" />
              </svg>
              <span>批量翻译</span>
            </button>
            <button class="action-btn" @click="exportXML">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
              </svg>
              <span>导出XML</span>
            </button>
            <button class="action-btn" @click="exportTSV">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                <polyline points="14 2 14 8 20 8" />
              </svg>
              <span>导出TSV</span>
            </button>
          </div>
        </div>
        <div class="import-table">
          <table>
            <thead>
              <tr>
                <th>KEY</th>
                <th>中文内容</th>
                <th>翻译结果</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(item, index) in importResult" :key="index">
                <td>{{ item.key }}</td>
                <td>{{ item.chinese }}</td>
                <td class="translation-cell">{{ item.translation || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      
      <div class="glossary-section">
        <div class="glossary-title">专业术语对照 (中文 → 英文)</div>
        <div class="glossary-list">
          <div v-for="item in getGlossaryList()" :key="item.zh" class="glossary-item">
            <span class="glossary-zh">{{ item.zh }}</span>
            <span class="glossary-arrow">→</span>
            <span class="glossary-en">{{ item.en }}</span>
          </div>
        </div>
      </div>
    </GlassCard>
  </div>
</template>

<style scoped>
.ai-translate {
  max-width: 100%;
}

.translate-section {
  display: flex;
  gap: 20px;
  align-items: flex-start;
  flex-wrap: wrap;
}

.input-panel,
.output-panel {
  flex: 1;
  min-width: 300px;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.panel-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.copy-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  background: rgba(22, 93, 255, 0.1);
  border: 1px solid rgba(22, 93, 255, 0.3);
  border-radius: 6px;
  color: #165DFF;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(22, 93, 255, 0.2);
  }
}

.text-input {
  width: 100%;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px;
  color: var(--text-primary);
  font-size: 14px;
  resize: vertical;
  outline: none;
  transition: border-color 0.2s;
  font-family: 'Inter', sans-serif;
  
  &:focus {
    border-color: rgba(22, 93, 255, 0.6);
  }
  
  &.output {
    background: rgba(0, 180, 42, 0.1);
    border-color: rgba(0, 180, 42, 0.3);
    color: #36CFC9;
  }
  
  &::placeholder {
    color: var(--text-tertiary);
  }
}

.swap-btn-wrap {
  flex-shrink: 0;
  padding-top: 40px;
}

.swap-btn {
  width: 48px;
  height: 48px;
  background: rgba(22, 93, 255, 0.2);
  border: 1px solid rgba(22, 93, 255, 0.4);
  border-radius: 50%;
  color: #165DFF;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(22, 93, 255, 0.3);
    transform: rotate(180deg);
  }
}

.action-buttons {
  display: flex;
  gap: 12px;
  margin-top: 20px;
  justify-content: center;
}

.translate-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 32px;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 10px;
  color: #FFFFFF;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.clear-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: rgba(100, 116, 139, 0.2);
  border: 1px solid rgba(100, 116, 139, 0.3);
  border-radius: 10px;
  color: #94A3B8;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(100, 116, 139, 0.3);
  }
}

.glossary-section {
  margin-top: 24px;
  padding: 16px;
  background: var(--bg-input);
  border-radius: 8px;
}

.glossary-title {
  font-size: 13px;
  color: var(--text-tertiary);
  margin-bottom: 12px;
}

.glossary-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.glossary-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(22, 93, 255, 0.08);
  border-radius: 6px;
}

.glossary-zh {
  font-size: 13px;
  color: #64748B;
}

.glossary-arrow {
  font-size: 12px;
  color: #64748B;
}

.glossary-en {
  font-size: 12px;
  color: #36CFC9;
}

@media screen and (max-width: 768px) {
  .translate-section {
    flex-direction: column;
    align-items: stretch;
  }
  
  .input-panel,
  .output-panel {
    min-width: 100%;
  }
  
  .swap-btn-wrap {
    padding-top: 16px;
    text-align: center;
  }
  
  .action-buttons {
    flex-direction: column;
    align-items: stretch;
  }
  
  .translate-btn,
  .clear-btn {
    justify-content: center;
  }
  
  .glossary-list {
    gap: 8px;
  }
}

@media screen and (max-width: 480px) {
  .text-input {
    font-size: 13px;
    padding: 10px;
  }
  
  .translate-btn,
  .clear-btn {
    padding: 10px 20px;
    font-size: 13px;
  }
  
  .glossary-item {
    padding: 6px 10px;
  }
  
  .glossary-zh {
    font-size: 12px;
  }
  
  .glossary-en {
    font-size: 11px;
  }
}
</style>
