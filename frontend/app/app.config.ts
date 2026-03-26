export default defineAppConfig({
  ui: {
    colors: {
      primary: 'skyblue', // 可切換: 'skyblue' | 'rosepink' | 'green' | 'blue' | 任何內建色
      secondary: 'rosepink',
      neutral: 'slate',
    },
    button: {
      base: 'cursor-pointer disabled:cursor-not-allowed'
    },
    dashboardSidebar: {
      slots: {
        root: 'bg-primary-100 dark:bg-primary-900 shadow-[4px_0_12px_rgba(0,0,0,0.15)] z-10',
      },
    },
    dashboardResizeHandle: {
      base: '',
    },
    dashboardNavbar: {
      slots: {
        root: 'bg-primary-50 dark:bg-primary-950',
      },
    },
    navigationMenu: {
      variants: {
        active: {
          false: {
            link: 'text-highlighted',
            linkLeadingIcon: 'text-default',
          },
        },
      },
    }
  }
})
