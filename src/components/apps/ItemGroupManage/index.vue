<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { FormInst, FormRules, SelectOption } from 'naive-ui'
import { NButton, NCard, NForm, NFormItem, NInput, NSelect, useDialog, useMessage } from 'naive-ui'
import { VueDraggable } from 'vue-draggable-plus'
import { deletes, edit, getList, saveSort } from '@/api/panel/itemIconGroup'
import { RoundCardModal, SvgIcon } from '@/components/common'
import { t } from '@/locales'
import { defaultNetworkConfig, genNetworkStateId } from '@/utils/network'

interface EditModalArg {
  show: boolean
  editStatus: number // 1.添加 2.编辑
  model: Panel.ItemIconGroup
  rules: FormRules
}

const formRef = ref<FormInst | null>(null)
const ms = useMessage()
const dialog = useDialog()
const sortStatus = ref(false)

const defaultMNodal: Panel.ItemIconGroup = {
  title: '',
  icon: 'material-symbols:folder-outline',
  sort: 9999,
  networkConfig: defaultNetworkConfig(),
}

const editModalArg = ref<EditModalArg>({
  show: false,
  editStatus: 1,
  model: { ...defaultMNodal, networkConfig: defaultNetworkConfig() },
  rules: {
    title: [
      {
        required: true,
        trigger: 'blur',
        message: t('form.required'),
      },
    ],
  },
})

const groups = ref<Panel.ItemIconGroup[]>([])

const typeOptions: SelectOption[] = [
  { label: t('apps.itemGroupManage.typeWan'), value: 'wan' },
  { label: t('apps.itemGroupManage.typeLan'), value: 'lan' },
  { label: t('apps.itemGroupManage.typeCustom'), value: 'custom' },
]

function ensureModelNetworkConfig() {
  if (!editModalArg.value.model.networkConfig)
    editModalArg.value.model.networkConfig = defaultNetworkConfig()
  if (!editModalArg.value.model.networkConfig.networkStates)
    editModalArg.value.model.networkConfig.networkStates = defaultNetworkConfig().networkStates
}

function handleAddGroup() {
  editModalArg.value.editStatus = 1
  editModalArg.value.model = { ...defaultMNodal, networkConfig: defaultNetworkConfig() }
  editModalArg.value.show = true
}

function handleEditGroup(groupInfo: Panel.ItemIconGroup) {
  editModalArg.value.show = true
  editModalArg.value.model = {
    ...groupInfo,
    networkConfig: groupInfo.networkConfig
      ? JSON.parse(JSON.stringify(groupInfo.networkConfig))
      : defaultNetworkConfig(),
  }
  editModalArg.value.editStatus = 2
  ensureModelNetworkConfig()
}

function handleDragSort() {
  sortStatus.value = true
}

function handleSaveSort() {
  const saveItems: Common.SortItemRequest[] = []
  for (let i = 0; i < groups.value.length; i++) {
    const element = groups.value[i]
    saveItems.push({
      id: element.id as number,
      sort: i + 1,
    })
  }
  saveSort(saveItems).then(({ code, msg }) => {
    if (code === 0) {
      ms.success(t('common.saveSuccess'))
      sortStatus.value = false
    }
    else {
      ms.error(`${t('common.saveFail')}:${msg}`)
    }
  })
}

function handleDelete(groupInfo: Panel.ItemIconGroup) {
  dialog.warning({
    title: t('common.warning'),
    content: t('apps.itemGroupManage.deleteWarnText', { name: groupInfo.title }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => {
      if (groupInfo.id) {
        deletes([groupInfo.id]).then(({ code, msg }) => {
          if (code !== 0)
            ms.error(t('common.deleteFail'))
          else
            refreshList()
        })
      }
    },

  })
}

function handleAddNetworkState() {
  ensureModelNetworkConfig()
  const cfg = editModalArg.value.model.networkConfig!
  cfg.networkStates.push({
    id: genNetworkStateId(),
    name: t('apps.itemGroupManage.typeCustom'),
    type: 'custom',
    host: '',
  })
}

function handleRemoveNetworkState(index: number) {
  ensureModelNetworkConfig()
  const cfg = editModalArg.value.model.networkConfig!
  if (cfg.networkStates.length <= 1) {
    ms.warning(t('common.warning'))
    return
  }
  const removed = cfg.networkStates.splice(index, 1)[0]
  if (cfg.activeNetworkStateId === removed.id)
    cfg.activeNetworkStateId = cfg.networkStates[0].id
}

function handleSaveGroup() {
  formRef.value?.validate((errors) => {
    if (!errors) {
      ensureModelNetworkConfig()
      const cfg = editModalArg.value.model.networkConfig!
      // Trim hosts so bare IPs like "192.168.1.10" persist cleanly
      for (const state of cfg.networkStates)
        state.host = (state.host || '').trim()

      if (!cfg.activeNetworkStateId || !cfg.networkStates.find(s => s.id === cfg.activeNetworkStateId))
        cfg.activeNetworkStateId = cfg.networkStates[0]?.id || 'wan'

      edit(editModalArg.value.model).then(({ code, msg, data }) => {
        if (code !== 0) {
          ms.error(msg)
          return
        }

        ms.success(t('common.saveSuccess'))
        // Keep manage list in sync with what we just saved
        if (data) {
          const saved = data as Panel.ItemIconGroup
          const idx = groups.value.findIndex(g => g.id === saved.id)
          if (idx >= 0)
            groups.value[idx] = { ...groups.value[idx], ...saved }
          else if (editModalArg.value.editStatus === 1)
            refreshList()
        }
        else {
          refreshList()
        }

        // Notify home page to reload groups so jump URLs use the new hosts
        window.dispatchEvent(new CustomEvent('sunpanel:groups-changed'))

        editModalArg.value.show = false
        editModalArg.value.model = { ...defaultMNodal, networkConfig: defaultNetworkConfig() }
      })
    }
    else { console.log(errors) }
  })
}

function refreshList() {
  getList<Common.ListResponse<Panel.ItemIconGroup[]>>().then(({ code, data }) => {
    if (code === 0 && data?.list)
      groups.value = data.list
  }).catch(() => {})
}

onMounted(() => {
  refreshList()
})
</script>

<template>
  <div class="h-full">
    <div class="p-2">
      <NButton type="success" size="small" style="margin-right: 10px;" @click="handleAddGroup">
        {{ $t('common.add') }}
      </NButton>

      <NButton v-if="!sortStatus" size="small" @click="handleDragSort">
        {{ $t('common.sort') }}
      </NButton>

      <NButton v-else type="warning" size="small" @click="handleSaveSort">
        {{ $t('common.saveSort') }}
      </NButton>
    </div>

    <div class=" overflow-auto w-full mt-[20px]  bg-slate-200 dark:bg-zinc-900 rounded-xl" style="height:calc(100% - 65px)">
      <VueDraggable
        v-model="groups"
        item-key="sort" :animation="300"
        :style="{ padding: sortStatus ? '20px' : '10px' }"
        :disabled="!sortStatus"
      >
        <div v-for="(item, index) in groups" :key="index" class="w-full">
          <NCard size="small" style="border-radius:10px;margin-bottom: 10px;">
            <div class="flex" :class="sortStatus ? 'cursor-move' : ''">
              <div class="flex items-center">
                <span class="mr-[10px]">
                  <SvgIcon class="text-[20px]" icon="material-symbols:ad-group-outline-rounded" />
                </span>
                <span>
                  {{ item.title }}
                </span>
              </div>
              <div class="ml-auto">
                <span>
                  <NButton strong secondary type="success" size="small" @click="handleEditGroup(item)">
                    <template #icon>
                      <SvgIcon icon="basil:edit-solid" />
                    </template>
                  </NButton>
                </span>
                <span class="ml-[10px]">
                  <NButton strong secondary type="error" size="small" class="ml-[10px]" @click="handleDelete(item)">
                    <template #icon>
                      <SvgIcon icon="material-symbols:delete" />
                    </template>
                  </NButton>
                </span>
              </div>
            </div>
          </NCard>
        </div>
      </VueDraggable>
    </div>

    <RoundCardModal v-model:show="editModalArg.show" size="small" type="small" :title="editModalArg.editStatus === 1 ? '添加' : '编辑'" style="width: 520px;">
      <NForm ref="formRef" :model="editModalArg.model" :rules="editModalArg.rules">
        <NFormItem path="title" :label="$t('apps.itemGroupManage.groupName')">
          <NInput v-model:value="editModalArg.model.title" type="text" :maxlength="20" show-count />
        </NFormItem>

        <div class="mb-2 font-medium">
          {{ $t('apps.itemGroupManage.networkStates') }}
        </div>

        <div
          v-for="(state, sIndex) in editModalArg.model.networkConfig?.networkStates || []"
          :key="state.id"
          class="mb-3 p-3 rounded-lg bg-slate-100 dark:bg-zinc-800"
        >
          <div class="flex gap-2 mb-2">
            <NInput
              v-model:value="state.name"
              :placeholder="$t('apps.itemGroupManage.networkStateName')"
              class="flex-1"
            />
            <NSelect
              v-model:value="state.type"
              :options="typeOptions"
              style="width: 110px"
            />
            <NButton
              strong
              secondary
              type="error"
              size="small"
              :disabled="(editModalArg.model.networkConfig?.networkStates.length || 0) <= 1"
              @click="handleRemoveNetworkState(sIndex)"
            >
              {{ $t('apps.itemGroupManage.deleteNetworkState') }}
            </NButton>
          </div>
          <NInput
            v-if="state.type !== 'wan'"
            v-model:value="state.host"
            :placeholder="$t('apps.itemGroupManage.networkStateHostPlaceholder')"
          />
        </div>

        <NButton size="small" type="info" secondary class="mb-3" @click="handleAddNetworkState">
          {{ $t('apps.itemGroupManage.addNetworkState') }}
        </NButton>

        <NFormItem :label="$t('apps.itemGroupManage.activeNetworkState')">
          <NSelect
            v-model:value="editModalArg.model.networkConfig!.activeNetworkStateId"
            :options="(editModalArg.model.networkConfig?.networkStates || []).map(s => ({ label: s.name, value: s.id }))"
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <NButton type="success" size="small" class="float-right" @click="handleSaveGroup">
          {{ $t('common.confirm') }}
        </NButton>
      </template>
    </RoundCardModal>
  </div>
</template>
