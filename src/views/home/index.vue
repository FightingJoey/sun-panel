<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import { NBackTop, NButton, NButtonGroup, NDropdown, NModal, NSkeleton, NSpin, useDialog, useMessage } from 'naive-ui'
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { AppIcon, AppStarter, EditItem } from './components'
import { Clock, SearchBox, SystemMonitor } from '@/components/deskModule'
import { SvgIcon } from '@/components/common'
import { deletes, getListByGroupId, saveSort } from '@/api/panel/itemIcon'
import { getList as getGroupList, switchNetworkState } from '@/api/panel/itemIconGroup'

import { setTitle, updateLocalUserInfo } from '@/utils/cmn'
import { useAuthStore, usePanelState } from '@/store'
import { PanelPanelConfigStyleEnum } from '@/enums'
import { VisitMode } from '@/enums/auth'
import { router } from '@/router'
import { t } from '@/locales'
import { buildJumpUrl, buildUrlForState, ensureNetworkConfig, getActiveNetworkState } from '@/utils/network'

interface ItemGroup extends Panel.ItemIconGroup {
  sortStatus?: boolean
  hoverStatus: boolean
  items?: Panel.ItemInfo[]
}

const ms = useMessage()
const dialog = useDialog()
const panelState = usePanelState()
const authStore = useAuthStore()

const scrollContainerRef = ref<HTMLElement | undefined>(undefined)

const editItemInfoShow = ref<boolean>(false)
const editItemInfoData = ref<Panel.ItemInfo | null>(null)
const windowShow = ref<boolean>(false)
const windowSrc = ref<string>('')
const windowTitle = ref<string>('')

const windowIframeRef = ref(null)
const windowIframeIsLoad = ref<boolean>(false)

const dropdownMenuX = ref(0)
const dropdownMenuY = ref(0)
const dropdownShow = ref(false)
const currentRightSelectItem = ref<Panel.ItemInfo | null>(null)
const currentRightSelectGroup = ref<ItemGroup | null>(null)
const currentAddItenIconGroupId = ref<number | undefined>()

const settingModalShow = ref(false)

const items = ref<ItemGroup[]>([])
const filterItems = ref<ItemGroup[]>([])
const globalSortMode = ref(false)

function openPage(openMethod: number, url: string, title?: string) {
  if (!url)
    return
  switch (openMethod) {
    case 1:
      window.location.href = url
      break
    case 2:
      window.open(url)
      break
    case 3:
      windowShow.value = true
      windowSrc.value = url
      windowTitle.value = title || url
      windowIframeIsLoad.value = true
      break

    default:
      break
  }
}

function findGroupForItem(item: Panel.ItemInfo): ItemGroup | undefined {
  const groupId = item.itemIconGroupId
  if (groupId)
    return items.value.find(g => g.id === groupId)
  return items.value.find(g => g.items?.some(i => i.id === item.id))
}

function resolveJumpUrl(item: Panel.ItemInfo, group?: ItemGroup | null): string {
  const g = group || findGroupForItem(item)
  return buildJumpUrl(item, g)
}

function handleItemClick(itemGroup: ItemGroup, item: Panel.ItemInfo) {
  if (globalSortMode.value || itemGroup.sortStatus) {
    handleEditItem(item)
    return
  }

  // Use the group object from the rendered list (filterItems) so the active
  // network state matches what the user just switched — index into `items`
  // can diverge after search filtering.
  const group = itemGroup || findGroupForItem(item)
  const jumpUrl = resolveJumpUrl(item, group)
  openPage(item.openMethod, jumpUrl, item.title)
}

function handWindowIframeIdLoad(payload: Event) {
  windowIframeIsLoad.value = false
}

function getList() {
  // 获取组数据
  getGroupList<Common.ListResponse<ItemGroup[]>>().then(({ code, data, msg }) => {
    if (code === 0 && data?.list) {
      items.value = data.list.map(g => ({
        ...g,
        hoverStatus: false,
        sortStatus: globalSortMode.value,
        networkConfig: JSON.parse(JSON.stringify(ensureNetworkConfig(g))),
        items: g.items || [],
      }))
      for (let i = 0; i < items.value.length; i++) {
        const element = items.value[i]
        if (element.id)
          updateItemIconGroupByNet(i, element.id)
      }
      filterItems.value = items.value
    }
    else if (code !== 0) {
      console.error('getGroupList failed', code, msg)
    }
  }).catch((err) => {
    console.error('getGroupList error', err)
  })
}

// 从后端获取组下面的图标
function updateItemIconGroupByNet(itemIconGroupIndex: number, itemIconGroupId: number) {
  getListByGroupId<Common.ListResponse<Panel.ItemInfo[]>>(itemIconGroupId).then((res) => {
    if (res.code === 0)
      items.value[itemIconGroupIndex].items = res.data.list
  })
}

function handleRightMenuSelect(key: string | number) {
  dropdownShow.value = false
  const item = currentRightSelectItem.value
  const group = currentRightSelectGroup.value
  if (!item)
    return

  if (typeof key === 'string' && key.startsWith('openState:')) {
    const stateId = key.slice('openState:'.length)
    const cfg = ensureNetworkConfig(group)
    const state = cfg.networkStates.find(s => s.id === stateId)
    if (state) {
      const url = buildUrlForState(item, state)
      openPage(item.openMethod, url, item.title)
    }
    return
  }

  switch (key) {
    case 'newWindows':
      window.open(resolveJumpUrl(item, group))
      break
    case 'openWanUrl':
      openPage(item.openMethod, item.url, item.title)
      break
    case 'edit':
      handleEditItem({ ...item })
      break
    case 'delete':
      dialog.warning({
        title: t('common.warning'),
        content: t('common.deleteConfirmByName', { name: item.title }),
        positiveText: t('common.confirm'),
        negativeText: t('common.cancel'),
        onPositiveClick: () => {
          deletes([item.id as number]).then(({ code, msg }) => {
            if (code === 0) {
              ms.success(t('common.deleteSuccess'))
              getList()
            }
            else {
              ms.error(`${t('common.deleteFail')}:${msg}`)
            }
          })
        },
      })

      break
    default:
      break
  }
}

function handleContextMenu(e: MouseEvent, itemGroup: ItemGroup, item: Panel.ItemInfo) {
  if (globalSortMode.value || itemGroup.sortStatus)
    return

  e.preventDefault()
  currentRightSelectItem.value = item
  currentRightSelectGroup.value = itemGroup || null
  dropdownShow.value = false
  nextTick().then(() => {
    dropdownShow.value = true
    dropdownMenuX.value = e.clientX
    dropdownMenuY.value = e.clientY
  })
}

function onClickoutside() {
  dropdownShow.value = false
}

function handleEditSuccess(item: Panel.ItemInfo) {
  getList()
}

function handleSaveSort() {
  const saveItems: Panel.ItemIconSortItem[] = []
  for (const group of items.value) {
    if (!group.items || !group.id)
      continue
    for (let i = 0; i < group.items.length; i++) {
      const element = group.items[i]
      element.itemIconGroupId = group.id
      saveItems.push({
        id: element.id as number,
        sort: i + 1,
        itemIconGroupId: group.id,
      })
    }
  }

  saveSort({ sortItems: saveItems }).then(({ code, msg }) => {
    if (code === 0) {
      ms.success(t('common.saveSuccess'))
      setGlobalSortMode(false)
    }
    else {
      ms.error(`${t('common.saveFail')}:${msg}`)
    }
  })
}

function getDropdownMenuOptions() {
  const dropdownMenuOptions: { label: string; key: string }[] = [
    {
      label: t('iconItem.newWindowOpen'),
      key: 'newWindows',
    },
  ]

  const group = currentRightSelectGroup.value
  const item = currentRightSelectItem.value
  if (group && item) {
    const cfg = ensureNetworkConfig(group)
    const active = getActiveNetworkState(group)
    for (const state of cfg.networkStates) {
      if (state.id === active?.id)
        continue
      dropdownMenuOptions.push({
        label: t('panelHome.openWithState', { name: state.name }),
        key: `openState:${state.id}`,
      })
    }
    if (active?.type !== 'wan') {
      dropdownMenuOptions.push({
        label: t('panelHome.openPublicUrl'),
        key: 'openWanUrl',
      })
    }
  }

  if (authStore.visitMode === VisitMode.VISIT_MODE_LOGIN) {
    dropdownMenuOptions.push({
      label: t('common.edit'),
      key: 'edit',
    }, {
      label: t('common.delete'),
      key: 'delete',
    })
  }

  return dropdownMenuOptions
}

function handleGroupsChanged() {
  getList()
}

onMounted(() => {
  // 更新用户信息
  updateLocalUserInfo()
  getList()

  // 更新同步云端配置
  panelState.updatePanelConfigByCloud()

  // 设置标题
  if (panelState.panelConfig.logoText)
    setTitle(panelState.panelConfig.logoText)

  window.addEventListener('sunpanel:groups-changed', handleGroupsChanged)
})

onUnmounted(() => {
  window.removeEventListener('sunpanel:groups-changed', handleGroupsChanged)
})

// Reload groups when the app launcher closes (e.g. after editing group hosts)
watch(settingModalShow, (visible, wasVisible) => {
  if (wasVisible && !visible)
    getList()
})

// 前端搜索过滤
function itemFrontEndSearch(keyword?: string) {
  keyword = keyword?.trim()
  if (keyword !== '' && panelState.panelConfig.searchBoxSearchIcon) {
    const filteredData = ref<ItemGroup[]>([])
    for (let i = 0; i < items.value.length; i++) {
      const element = items.value[i].items?.filter((item: Panel.ItemInfo) => {
        return (
          item.title.toLowerCase().includes(keyword?.toLowerCase() ?? '')
          || item.url.toLowerCase().includes(keyword?.toLowerCase() ?? '')
          || item.description?.toLowerCase().includes(keyword?.toLowerCase() ?? '')
        )
      })
      if (element && element.length > 0)
        filteredData.value.push({ ...items.value[i], items: element, hoverStatus: false })
    }
    filterItems.value = filteredData.value
  }
  else {
    filterItems.value = items.value
  }
}

function handleSetHoverStatus(groupIndex: number, hoverStatus: boolean) {
  if (items.value[groupIndex])
    items.value[groupIndex].hoverStatus = hoverStatus
}

function setGlobalSortMode(enabled: boolean) {
  globalSortMode.value = enabled
  for (const g of items.value)
    g.sortStatus = enabled

  if (!enabled)
    getList()
}

function handleToggleSortMode() {
  if (globalSortMode.value)
    setGlobalSortMode(false)
  else
    setGlobalSortMode(true)
}

function handleEditItem(item: Panel.ItemInfo) {
  editItemInfoData.value = item
  editItemInfoShow.value = true
  currentAddItenIconGroupId.value = undefined
}

function handleAddItem(itemIconGroupId?: number) {
  editItemInfoData.value = null
  editItemInfoShow.value = true
  if (itemIconGroupId)
    currentAddItenIconGroupId.value = itemIconGroupId
}

function getNetworkStateOptions(group: ItemGroup) {
  const cfg = ensureNetworkConfig(group)
  return cfg.networkStates.map(s => ({
    label: s.name,
    key: s.id,
  }))
}

function syncGroupNetworkConfig(groupId: number, cfg: Panel.NetworkConfig) {
  const cloned = JSON.parse(JSON.stringify(cfg)) as Panel.NetworkConfig
  const apply = (list: ItemGroup[]) => {
    const g = list.find(x => x.id === groupId)
    if (g)
      g.networkConfig = cloned
  }
  apply(items.value)
  if (filterItems.value !== items.value)
    apply(filterItems.value)
}

function handleSwitchNetworkState(group: ItemGroup, stateId: string | number) {
  const cfg = ensureNetworkConfig(group)
  const id = String(stateId)
  const state = cfg.networkStates.find(s => s.id === id)
  if (!state || !group.id)
    return

  // Optimistic UI (deep clone so we never mutate a shared stale object)
  const next = JSON.parse(JSON.stringify(cfg)) as Panel.NetworkConfig
  next.activeNetworkStateId = id
  syncGroupNetworkConfig(group.id, next)

  // Backend only flips activeNetworkStateId on the DB row — never resends hosts,
  // so a stale home-page config cannot wipe a host saved in 分组管理.
  switchNetworkState<Panel.ItemIconGroup>(group.id, id).then(({ code, msg, data }) => {
    if (code === 0) {
      if (data?.networkConfig)
        syncGroupNetworkConfig(group.id, data.networkConfig)
      ms.success(t('panelHome.switchNetworkStateSuccess', { name: state.name }))
    }
    else {
      ms.error(`${t('common.saveFail')}:${msg}`)
      getList()
    }
  }).catch(() => {
    getList()
  })
}

function activeNetworkStateName(group: ItemGroup): string {
  return getActiveNetworkState(group)?.name || ''
}
</script>

<template>
  <div class="w-full h-full sun-main">
    <div
      class="cover wallpaper" :style="{
        filter: `blur(${panelState.panelConfig.backgroundBlur}px)`,
        background: `url(${panelState.panelConfig.backgroundImageSrc}) no-repeat`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }"
    />
    <div class="mask" :style="{ backgroundColor: `rgba(0,0,0,${panelState.panelConfig.backgroundMaskNumber})` }" />
    <div ref="scrollContainerRef" class="absolute w-full h-full overflow-auto">
      <div
        class="p-2.5 mx-auto"
        :style="{
          marginTop: `${panelState.panelConfig.marginTop}%`,
          marginBottom: `${panelState.panelConfig.marginBottom}%`,
          maxWidth: (panelState.panelConfig.maxWidth ?? '1200') + panelState.panelConfig.maxWidthUnit,
        }"
      >
        <!-- 头 -->
        <div class="mx-[auto] w-[80%]">
          <div class="flex mx-[auto] items-center justify-center text-white">
            <div class="logo">
              <span class="text-2xl md:text-6xl font-bold text-shadow">
                {{ panelState.panelConfig.logoText }}
              </span>
            </div>
            <div class="divider text-base lg:text-2xl mx-[10px]">
              |
            </div>
            <div class="text-shadow">
              <Clock :hide-second="!panelState.panelConfig.clockShowSecond" />
            </div>
          </div>
          <div v-if="panelState.panelConfig.searchBoxShow" class="flex mt-[20px] mx-auto sm:w-full lg:w-[80%]">
            <SearchBox @itemSearch="itemFrontEndSearch" />
          </div>
        </div>

        <!-- 应用盒子 -->
        <div :style="{ marginLeft: `${panelState.panelConfig.marginX}px`, marginRight: `${panelState.panelConfig.marginX}px` }">
          <!-- 系统监控状态 -->
          <div
            v-if="panelState.panelConfig.systemMonitorShow
              && ((panelState.panelConfig.systemMonitorPublicVisitModeShow && authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC)
                || authStore.visitMode === VisitMode.VISIT_MODE_LOGIN)"
            class="flex mx-auto"
          >
            <SystemMonitor
              :allow-edit="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN"
              :show-title="panelState.panelConfig.systemMonitorShowTitle"
            />
          </div>

          <!-- 全局排序工具栏 -->
          <div
            v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN && globalSortMode"
            class="mt-[20px] flex items-center gap-2"
          >
            <NButton color="#2a2a2a6b" @click="handleSaveSort">
              <template #icon>
                <SvgIcon class="text-white font-xl" icon="material-symbols:save" />
              </template>
              <div>
                {{ $t('panelHome.saveGlobalSort') }}
              </div>
            </NButton>
            <NButton color="#2a2a2a6b" @click="setGlobalSortMode(false)">
              {{ $t('common.cancel') }}
            </NButton>
          </div>

          <!-- 组纵向排列 -->
          <div
            v-for="(itemGroup, itemGroupIndex) in filterItems" :key="itemGroup.id || itemGroupIndex"
            class="item-list mt-[50px]"
            :class="globalSortMode ? 'shadow-2xl border shadow-[0_0_30px_10px_rgba(0,0,0,0.3)]  p-[10px] rounded-2xl' : ''"
            @mouseenter="handleSetHoverStatus(itemGroupIndex, true)"
            @mouseleave="handleSetHoverStatus(itemGroupIndex, false)"
          >
            <!-- 分组标题 -->
            <div class="text-white text-xl font-extrabold mb-[20px] ml-[10px] flex items-center flex-wrap gap-2">
              <span class="group-title text-shadow">
                {{ itemGroup.title }}
              </span>

              <!-- 分组级网络状态切换 -->
              <NDropdown
                trigger="click"
                :options="getNetworkStateOptions(itemGroup)"
                @select="(key) => handleSwitchNetworkState(itemGroup, key)"
              >
                <span
                  class="text-sm font-normal px-2 py-0.5 rounded-full cursor-pointer bg-black/30 hover:bg-black/50 text-shadow"
                  :title="t('panelHome.switchNetworkState')"
                >
                  {{ activeNetworkStateName(itemGroup) || t('panelHome.switchNetworkState') }}
                </span>
              </NDropdown>

              <div
                v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN"
                class="group-buttons ml-2 delay-100 transition-opacity flex"
                :class="itemGroup.hoverStatus || globalSortMode ? 'opacity-100' : 'opacity-0'"
              >
                <span class="mr-2 cursor-pointer" :title="t('common.add')" @click="handleAddItem(itemGroup.id)">
                  <SvgIcon class="text-white font-xl" icon="typcn:plus" />
                </span>
                <span
                  class="mr-2 cursor-pointer"
                  :title="t('panelHome.globalSort')"
                  @click="handleToggleSortMode"
                >
                  <SvgIcon class="text-white font-xl" icon="ri:drag-drop-line" />
                </span>
              </div>
            </div>

            <!-- 详情图标 -->
            <div v-if="panelState.panelConfig.iconStyle === PanelPanelConfigStyleEnum.info">
              <div v-if="itemGroup.items">
                <VueDraggable
                  v-model="itemGroup.items" item-key="id" :animation="300"
                  class="icon-info-box"
                  filter=".not-drag"
                  :group="{ name: 'panel-icons', pull: true, put: true }"
                  :disabled="!globalSortMode"
                >
                  <div v-for="item in itemGroup.items" :key="item.id" :title="item.description" @contextmenu="(e) => handleContextMenu(e, itemGroup, item)">
                    <AppIcon
                      :class="globalSortMode ? 'cursor-move' : 'cursor-pointer'"
                      :item-info="item"
                      :icon-text-color="panelState.panelConfig.iconTextColor"
                      :icon-text-info-hide-description="panelState.panelConfig.iconTextInfoHideDescription || false"
                      :icon-text-icon-hide-title="panelState.panelConfig.iconTextIconHideTitle || false"
                      :style="0"
                      @click="handleItemClick(itemGroup, item)"
                    />
                  </div>

                  <div v-if="itemGroup.items.length === 0" class="not-drag">
                    <AppIcon
                      :class="globalSortMode ? 'cursor-move' : 'cursor-pointer'"
                      :item-info="{ icon: { itemType: 3, text: 'subway:add' }, title: t('common.add'), url: '', openMethod: 0 }"
                      :icon-text-color="panelState.panelConfig.iconTextColor"
                      :icon-text-info-hide-description="panelState.panelConfig.iconTextInfoHideDescription || false"
                      :icon-text-icon-hide-title="panelState.panelConfig.iconTextIconHideTitle || false"
                      :style="0"
                      @click="handleAddItem(itemGroup.id)"
                    />
                  </div>
                </VueDraggable>
              </div>
            </div>

            <!-- APP图标宫型盒子 -->
            <div v-if="panelState.panelConfig.iconStyle === PanelPanelConfigStyleEnum.icon">
              <div v-if="itemGroup.items">
                <VueDraggable
                  v-model="itemGroup.items" item-key="id" :animation="300"
                  class="icon-small-box"
                  filter=".not-drag"
                  :group="{ name: 'panel-icons', pull: true, put: true }"
                  :disabled="!globalSortMode"
                >
                  <div v-for="item in itemGroup.items" :key="item.id" :title="item.description" @contextmenu="(e) => handleContextMenu(e, itemGroup, item)">
                    <AppIcon
                      :class="globalSortMode ? 'cursor-move' : 'cursor-pointer'"
                      :item-info="item"
                      :icon-text-color="panelState.panelConfig.iconTextColor"
                      :icon-text-info-hide-description="!panelState.panelConfig.iconTextInfoHideDescription"
                      :icon-text-icon-hide-title="panelState.panelConfig.iconTextIconHideTitle || false"
                      :style="1"
                      @click="handleItemClick(itemGroup, item)"
                    />
                  </div>

                  <div v-if="itemGroup.items.length === 0" class="not-drag">
                    <AppIcon
                      class="cursor-pointer"
                      :item-info="{ icon: { itemType: 3, text: 'subway:add' }, title: $t('common.add'), url: '', openMethod: 0 }"
                      :icon-text-color="panelState.panelConfig.iconTextColor"
                      :icon-text-info-hide-description="!panelState.panelConfig.iconTextInfoHideDescription"
                      :icon-text-icon-hide-title="panelState.panelConfig.iconTextIconHideTitle || false"
                      :style="1"
                      @click="handleAddItem(itemGroup.id)"
                    />
                  </div>
                </VueDraggable>
              </div>
            </div>
          </div>
        </div>
        <div class="mt-5 footer" v-html="panelState.panelConfig.footerHtml" />
      </div>
    </div>

    <!-- 右键菜单 -->
    <NDropdown
      placement="bottom-start" trigger="manual" :x="dropdownMenuX" :y="dropdownMenuY"
      :options="getDropdownMenuOptions()" :show="dropdownShow" :on-clickoutside="onClickoutside" @select="handleRightMenuSelect"
    />

    <!-- 悬浮按钮 -->
    <div class="fixed-element shadow-[0_0_10px_2px_rgba(0,0,0,0.2)]">
      <NButtonGroup vertical>
        <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" color="#2a2a2a6b" @click="settingModalShow = !settingModalShow">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="majesticons-applications" />
          </template>
        </NButton>

        <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC" color="#2a2a2a6b" :title="$t('panelHome.goToLogin')" @click="router.push('/login')">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="material-symbols:account-circle" />
          </template>
        </NButton>
      </NButtonGroup>

      <AppStarter v-model:visible="settingModalShow" />
    </div>

    <NBackTop
      :listen-to="() => scrollContainerRef"
      :right="10"
      :bottom="10"
      style="background-color:transparent;border: none;box-shadow: none;"
    >
      <div class="shadow-[0_0_10px_2px_rgba(0,0,0,0.2)]">
        <NButton color="#2a2a2a6b">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="icon-park-outline:to-top" />
          </template>
        </NButton>
      </div>
    </NBackTop>

    <EditItem v-model:visible="editItemInfoShow" :item-info="editItemInfoData" :item-group-id="currentAddItenIconGroupId" @done="handleEditSuccess" />

    <!-- 弹窗 -->
    <NModal
      v-model:show="windowShow" :mask-closable="false" preset="card"
      style="max-width: 1000px;height: 600px;border-radius: 1rem;" :bordered="true" size="small" role="dialog"
      aria-modal="true"
    >
      <template #header>
        <div class="flex items-center">
          <span class="mr-[20px]">
            {{ windowTitle }}
          </span>

          <NSpin v-if="windowIframeIsLoad" size="small" />
        </div>
      </template>
      <div class="w-full h-full rounded-2xl overflow-hidden border dark:border-zinc-700">
        <div v-if="windowIframeIsLoad" class="flex flex-col p-5">
          <NSkeleton height="50px" width="100%" class="rounded-lg" />
          <NSkeleton height="180px" width="100%" class="mt-[20px] rounded-lg" />
          <NSkeleton height="180px" width="100%" class="mt-[20px] rounded-lg" />
        </div>
        <iframe
          v-show="!windowIframeIsLoad" id="windowIframeId" ref="windowIframeRef" :src="windowSrc"
          class="w-full h-full" frameborder="0" @load="handWindowIframeIdLoad"
        />
      </div>
    </NModal>
  </div>
</template>

<style>
body,
html {
  overflow: hidden;
  background-color: rgb(54, 54, 54);
}
</style>

<style scoped>
.mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.sun-main {
  user-select: none;
}

.cover {
  position: absolute;
  width: 100%;
  height: 100%;
  overflow: hidden;

  transform: scale(1.05);
}

.text-shadow {
  text-shadow: 2px 2px 50px rgb(0, 0, 0);
}

.app-icon-text-shadow {
  text-shadow: 2px 2px 5px rgb(0, 0, 0);
}

.fixed-element {
  position: fixed;
  right: 10px;
  bottom: 50px;
}

.icon-info-box {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 18px;

}

.icon-small-box {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(75px, 1fr));
  gap: 18px;

}

@media (max-width: 500px) {
  .icon-info-box{
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  }
}
</style>
