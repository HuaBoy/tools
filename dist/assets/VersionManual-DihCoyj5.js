import{_ as k,a as p,c as m,p as V,q as h,b as e,e as B,M as w,F as Q,k as y,v as L,L as A,t as C,r as d,E as r,l as M}from"./index-M_yJGbPD.js";import{G as b}from"./GlassCard-B6xFHoUW.js";import{u as D}from"./logs-CQuUDDM3.js";const S={class:"version-manual"},E={class:"toolbar"},U={class:"version-select"},I=["value"],$={class:"search-wrap"},j=["disabled"],R={class:"manual-content"},F={__name:"VersionManual",setup(O){const i=D(),a=d("QB-2024-V1.0"),l=d(""),u=d(!1),f=[{value:"QB-2024-V1.0",label:"QB-2024-V1.0 标准版"},{value:"QB-2024-V1.1",label:"QB-2024-V1.1 增强版"},{value:"QB-2025-V2.0",label:"QB-2025-V2.0 专业版"},{value:"QB-2025-V2.1",label:"QB-2025-V2.1 旗舰版"}],x={"QB-2024-V1.0":`起爆器版本手册 - QB-2024-V1.0

一、产品概述
本产品为工业起爆器控制系统，用于精确控制爆破作业中的起爆时序和能量输出。

二、技术参数
- 工作电压: 12V DC
- 最大输出电流: 5A
- 起爆通道数: 4通道
- 延时精度: ±1ms
- 工作温度: -20°C ~ +50°C

三、安装说明
1. 将起爆器固定在安全位置
2. 连接电源线和通信线
3. 安装天线并调整位置
4. 连接雷管引线

四、操作指南
1. 开机自检
2. 参数配置
3. 起爆测试
4. 执行起爆

五、故障排除
见知识库故障代码说明`,"QB-2024-V1.1":`起爆器版本手册 - QB-2024-V1.1

一、产品概述
增强版在标准版基础上增加了远程控制功能和数据加密传输。

二、技术参数
- 工作电压: 12V DC
- 最大输出电流: 8A
- 起爆通道数: 8通道
- 延时精度: ±0.5ms
- 工作温度: -30°C ~ +60°C
- 通信距离: 5km

三、新增功能
- 远程无线控制
- AES-256数据加密
- 实时状态监控
- 自动故障诊断

四、安装说明
1. 安装主控单元
2. 配置无线模块
3. 设置加密密钥
4. 连接传感器

五、操作指南
详见标准版操作指南`,"QB-2025-V2.0":`起爆器版本手册 - QB-2025-V2.0

一、产品概述
专业版集成AI智能分析功能，支持自动参数优化和故障预测。

二、技术参数
- 工作电压: 24V DC
- 最大输出电流: 10A
- 起爆通道数: 16通道
- 延时精度: ±0.1ms
- 工作温度: -40°C ~ +70°C
- 通信距离: 10km

三、AI功能特性
- 智能参数优化
- 故障预测预警
- 数据分析报告
- 远程诊断支持

四、系统要求
- 配套AI分析平台
- 高速网络连接
- 专业培训认证

五、安全注意事项
严格遵守爆破安全规范`,"QB-2025-V2.1":`起爆器版本手册 - QB-2025-V2.1

一、产品概述
旗舰版是本系列最高配置产品，支持全自动化起爆和智能决策系统。

二、技术参数
- 工作电压: 24V DC
- 最大输出电流: 15A
- 起爆通道数: 32通道
- 延时精度: ±0.05ms
- 工作温度: -50°C ~ +80°C
- 通信距离: 20km

三、旗舰功能
- 全自动起爆流程
- 智能决策系统
- 多设备协同控制
- 三维可视化监控
- 应急预案自动触发

四、高级配置
详见技术白皮书和专业培训文档

五、维护保养
定期校准和软件更新`},n=M(()=>x[a.value]||""),_=()=>{i.addLog("切换","版本手册",`切换版本: ${a.value}`)},v=async()=>{if(!l.value.trim()){r.warning("请输入搜索内容");return}u.value=!0,await new Promise(s=>setTimeout(s,500)),u.value=!1;const o=n.value.toLowerCase().indexOf(l.value.toLowerCase());o!==-1?r.success(`找到匹配内容，位于第 ${n.value.substring(0,o).split(`
`).length} 行`):r.info("未找到匹配内容"),i.addLog("搜索","版本手册",`搜索: ${l.value}`)},g=()=>{const o=n.value,s=new Blob([o],{type:"text/plain"}),t=URL.createObjectURL(s),c=document.createElement("a");c.href=t,c.download=`manual_${a.value}.txt`,c.click(),URL.revokeObjectURL(t),i.addLog("导出","版本手册",`导出手册: ${a.value}`),r.success("手册导出成功")};return(o,s)=>(p(),m("div",S,[V(b,{title:"起爆器版本手册"},{default:h(()=>[e("div",E,[e("div",U,[s[2]||(s[2]=e("label",{class:"select-label"},"选择版本",-1)),B(e("select",{"onUpdate:modelValue":s[0]||(s[0]=t=>a.value=t),class:"version-dropdown",onChange:_},[(p(),m(Q,null,y(f,t=>e("option",{key:t.value,value:t.value},C(t.label),9,I)),64))],544),[[w,a.value]])]),e("div",$,[B(e("input",{"onUpdate:modelValue":s[1]||(s[1]=t=>l.value=t),type:"text",class:"search-input",placeholder:"全文搜索...",onKeyup:A(v,["enter"])},null,544),[[L,l.value]]),e("button",{class:"search-btn",disabled:u.value,onClick:v},[...s[3]||(s[3]=[e("svg",{width:"14",height:"14",viewBox:"0 0 24 24",fill:"none",stroke:"currentColor","stroke-width":"2","stroke-linecap":"round","stroke-linejoin":"round"},[e("circle",{cx:"11",cy:"11",r:"8"}),e("line",{x1:"21",y1:"21",x2:"16.65",y2:"16.65"})],-1),e("span",null,"全文AI检索",-1)])],8,j)]),e("button",{class:"export-btn",onClick:g},[...s[4]||(s[4]=[e("svg",{width:"14",height:"14",viewBox:"0 0 24 24",fill:"none",stroke:"currentColor","stroke-width":"2","stroke-linecap":"round","stroke-linejoin":"round"},[e("path",{d:"M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"}),e("polyline",{points:"7 10 12 15 17 10"}),e("line",{x1:"12",y1:"15",x2:"12",y2:"3"})],-1),e("span",null,"导出PDF手册",-1)])])])]),_:1}),V(b,{title:"手册内容",style:{"margin-top":"20px"}},{default:h(()=>[e("div",R,[e("pre",null,C(n.value),1)])]),_:1})]))}},P=k(F,[["__scopeId","data-v-4f6a7a9f"]]);export{P as default};
