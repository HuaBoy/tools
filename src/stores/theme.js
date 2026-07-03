import { defineStore } from 'pinia';

export const useThemeStore = defineStore('theme', {
  state: () => ({
    currentTheme: 'light'
  }),

  getters: {
    isLight: () => true,
    isDark: () => false,
    themeClass: () => 'theme-light'
  },

  actions: {
    toggleTheme() {},
    setTheme() {},
    applyTheme() {},
    saveTheme() {},
    loadTheme() {}
  }
});