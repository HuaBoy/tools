<script setup>import { ref, watch } from 'vue';
const props = defineProps({
 visible: {
 type: Boolean,
 default: false
 },
 title: {
 type: String,
 default: '确认操作'
 },
 content: {
 type: String,
 default: '确定要执行此操作吗？'
 },
 type: {
 type: String,
 default: 'warning'
 }
});
const emit = defineEmits(['confirm', 'cancel']);
const isVisible = ref(false);
watch(() => props.visible, (val) => {
 isVisible.value = val;
});
const handleConfirm = () => {
 emit('confirm');
};
const handleCancel = () => {
 emit('cancel');
};
</script>

<template>
  <Teleport to="body">
    <div v-if="isVisible" class="modal-overlay" @click.self="handleCancel">
      <div class="modal-content">
        <div class="modal-header">
          <span class="modal-title">{{ title }}</span>
          <button class="modal-close" @click="handleCancel">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="modal-icon" :class="type">
            <svg v-if="type === 'warning'" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
              <line x1="12" y1="9" x2="12" y2="13" />
              <line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            <svg v-else-if="type === 'success'" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
              <polyline points="22 4 12 14.01 9 11.01" />
            </svg>
            <svg v-else width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="15" y1="9" x2="9" y2="15" />
              <line x1="9" y1="9" x2="15" y2="15" />
            </svg>
          </div>
          <p class="modal-text">{{ content }}</p>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="handleCancel">取消</button>
          <button class="btn-confirm" :class="type" @click="handleConfirm">确定</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  backdrop-filter: blur(4px);
}

.modal-content {
  background: rgba(30, 41, 59, 0.95);
  border: 1px solid rgba(22, 93, 255, 0.3);
  border-radius: 12px;
  width: 400px;
  max-width: 90%;
  overflow: hidden;
  animation: modalIn 0.2s ease;
}

@keyframes modalIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(22, 93, 255, 0.2);
  background: rgba(22, 93, 255, 0.05);
}

.modal-title {
  font-size: 16px;
  font-weight: 600;
  color: #FFFFFF;
}

.modal-close {
  background: none;
  border: none;
  color: #64748B;
  cursor: pointer;
  padding: 4px;
  transition: color 0.2s;

  &:hover {
    color: #94A3B8;
  }
}

.modal-body {
  padding: 24px 20px;
  text-align: center;
}

.modal-icon {
  margin-bottom: 16px;

  &.warning {
    color: #FF7D00;
  }

  &.success {
    color: #00B42A;
  }

  &.error {
    color: #F53F3F;
  }
}

.modal-text {
  font-size: 14px;
  color: #94A3B8;
  line-height: 1.6;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid rgba(22, 93, 255, 0.2);
}

.btn-cancel {
  padding: 8px 20px;
  background: rgba(100, 116, 139, 0.2);
  border: 1px solid rgba(100, 116, 139, 0.3);
  border-radius: 8px;
  color: #94A3B8;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    background: rgba(100, 116, 139, 0.3);
  }
}

.btn-confirm {
  padding: 8px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;

  &.warning {
    background: #FF7D00;
    color: #fff;

    &:hover {
      background: #ff9533;
    }
  }

  &.success {
    background: #00B42A;
    color: #fff;

    &:hover {
      background: #00c832;
    }
  }

  &.error {
    background: #F53F3F;
    color: #fff;

    &:hover {
      background: #ff5a5a;
    }
  }
}
</style>
