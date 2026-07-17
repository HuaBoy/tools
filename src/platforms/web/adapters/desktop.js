// PC 端特有适配：键盘快捷键、右键菜单、粘贴板
export const desktopAdapter = {
  registerShortcut(key, ctrl, handler) {
    window.addEventListener('keydown', (e) => {
      if (e.key === key && e.ctrlKey === ctrl) { e.preventDefault(); handler() }
    })
  },
  copyToClipboard(text) {
    navigator.clipboard.writeText(text)
  }
}