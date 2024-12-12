<template>
  <component :is="iconComponent" />
</template>

<script setup>
import { computed } from 'vue'
import IconFlagUs from '~icons/circle-flags/us'
import IconFlagCn from '~icons/circle-flags/cn'

const props = defineProps({
  icon: {
    type: String,
    required: true
  }
})

const iconComponent = computed(() => {
  const icons = {
    'flag-us': IconFlagUs,
    'flag-cn': IconFlagCn
  }
  return icons[props.icon]
})
</script> 
