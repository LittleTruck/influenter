export default defineAppConfig({
  ui: {
    colors: {
      primary: 'iris', // 品牌紫（redesign.md #534AB7）。可切換: 'iris' | 'skyblue' | 'rosepink' | 任何內建色
      secondary: 'rosepink',
      neutral: 'slate',
    },
    button: {
      base: 'cursor-pointer disabled:cursor-not-allowed'
    },
    dashboardSidebar: {
      slots: {
        // Col 1｜漸層側欄（redesign.md）
        root: 'bg-[linear-gradient(160deg,#3C3489_0%,#534AB7_60%,#7F77DD_100%)] shadow-[4px_0_12px_rgba(83,74,183,0.25)] border-r-0 z-10',
        header: 'text-white',
        footer: 'text-white',
      },
    },
    dashboardResizeHandle: {
      base: '',
    },
    dashboardNavbar: {
      slots: {
        root: 'bg-primary-50 dark:bg-primary-950',
        title: 'text-xl',
      },
    },
    navigationMenu: {
      // Col 1｜漸層側欄上的導覽項目（白字 + 半透明 active pill）
      variants: {
        active: {
          true: {
            link: 'bg-white/18 !text-white font-medium before:hidden',
            linkLeadingIcon: '!text-white',
          },
          false: {
            link: '!text-white/55 hover:!text-white hover:bg-white/10',
            linkLeadingIcon: 'text-white/70 group-hover:text-white',
          },
        },
      },
    }
  }
})
