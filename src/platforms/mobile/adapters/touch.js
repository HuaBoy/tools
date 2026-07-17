// 移动端适配：触控手势、振动、扫码
export const touchAdapter = {
  vibrate(duration = 50) {
    navigator.vibrate?.(duration)
  },
  onSwipe(el, direction, handler) {
    let startX = 0, startY = 0
    el.addEventListener('touchstart', e => { startX = e.touches[0].clientX; startY = e.touches[0].clientY })
    el.addEventListener('touchend', e => {
      const dx = e.changedTouches[0].clientX - startX
      const dy = e.changedTouches[0].clientY - startY
      if (Math.abs(dx) > Math.abs(dy)) handler(dx > 50 ? 'right' : dx < -50 ? 'left' : null)
      else handler(dy > 50 ? 'down' : dy < -50 ? 'up' : null)
    })
  },
  vibrate: (d) => navigator.vibrate?.(d)
}