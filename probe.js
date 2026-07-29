const fs = require('fs');
const s = fs.readFileSync('E:/tools20260623/bigscreen_app.js', 'utf8');
const keys = ['password', 'login', 'autoLogin', 'type:"account"', 'grant_type', 'r()('];
keys.forEach(k => {
  let i = s.indexOf(k);
  while (i >= 0) {
    console.log('=== ' + k + ' @' + i + ' ===');
    console.log(s.slice(Math.max(0, i - 100), i + 100));
    console.log();
    i = s.indexOf(k, i + 1);
    if (i > 1e7) break;
  }
});
// 找调用登录函数的地方：搜索 .A)( 或导入 4253 后调用
const idx = s.indexOf('n(4253)');
console.log('=== import 4253 ===', idx);
if (idx >= 0) console.log(s.slice(idx - 120, idx + 120));
