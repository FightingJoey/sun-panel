declare namespace Panel {

    interface Info extends ItemInfo {

    }

    interface NetworkState {
        id: string
        name: string
        /** wan = public full URL; lan/custom = host + item.port */
        type: 'wan' | 'lan' | 'custom' | string
        host?: string
    }

    interface NetworkConfig {
        networkStates: NetworkState[]
        activeNetworkStateId: string
    }

    interface ItemInfo extends Common.InfoBase {
        icon: ItemIcon |null
        title: string
        url: string
        sort?: number
        lanUrl?: string
        port?: number | null
        description?: string
        openMethod: number
        itemIconGroupId ?:number
    }

    interface ItemIconGroup extends Common.InfoBase {
        icon?: string
        title?: string
        sort?:number
        description?: string
        networkConfig?: NetworkConfig
    }

    interface ItemIcon {
        itemType: number
        src ?: string
        text ?: string
        // bgColor ?: string
        backgroundColor ?: string
    }

    interface State {
        rightSiderCollapsed: boolean
        leftSiderCollapsed: boolean
        networkMode:PanelStateNetworkModeEnum | null
        panelConfig:panelConfig
    }

    interface panelConfig{
        backgroundImageSrc?:string
        backgroundBlur?:number
        backgroundMaskNumber?:number
        iconStyle?:PanelPanelConfigStyleEnum
        iconTextColor?:string
        iconTextInfoHideDescription?:boolean
        iconTextIconHideTitle?:boolean
        logoText?:string
        logoImageSrc?:string
        clockShowSecond?:boolean
        clockColor?:string
        searchBoxShow?:boolean
        searchBoxSearchIcon?:boolean
        marginTop?:number
        marginBottom?:number
        maxWidth?:number
        maxWidthUnit:string
        marginX?:number
        footerHtml?:string
        systemMonitorShow?:boolean
        systemMonitorShowTitle?:boolean
        systemMonitorPublicVisitModeShow?:boolean
        netModeChangeButtonShow?:boolean
    }

    interface userConfig{
        panel:panelConfig
        searchEngine?:any
    }

    interface ItemIconSortItem {
        id: number
        sort: number
        itemIconGroupId: number
    }

    interface ItemIconSortRequest{
        sortItems:ItemIconSortItem[]
        itemIconGroupId?:number
    }
}
