export default defineAppConfig({
  ui: {
    primary: 'green',
    gray: 'slate',
    button: {
      base: 'cursor-pointer disabled:cursor-not-allowed'
    },
    dashboardSidebar: {
      slots: {
        root: 'bg-gray-100 dark:bg-gray-950'
      }
    },
    dashboardNavbar: {
      slots: {
        root: 'bg-gray-50 dark:bg-gray-950'
      }
    }
  }
})


