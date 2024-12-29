import { css, definePreset } from '@primeng/themes';
import Aura from '@primeng/themes/aura';

export const FFXPreset = definePreset(Aura, {
    semantic: {
        primary: {
            50: '{sky.50}',
            100: '{sky.100}',
            200: '{sky.200}',
            300: '{sky.300}',
            400: '{sky.400}',
            500: '{sky.500}',
            600: '{sky.600}',
            700: '{sky.700}',
            800: '{sky.800}',
            900: '{sky.900}',
            950: '{sky.950}'
        },
    },
    components: {
        button: {
            colorScheme: {
                root: {
                    primary: {
                        background: 'transparent',
                        hoverBackground: '{sky.200}',
                        borderColor: 'transparent',
                        hoverBorderColor: '{sky.300}',
                        rippleOpacity: 0.2,
                        color: '{primary.600}',
                        hoverColor: '{primary.500}',
                    },
                },
            },
        },
        togglebutton: {
            content: {
                checkedShadow: '2px 2px 4px rgba(0, 0, 0, 0.2), -1px -1px 3px rgba(255, 255, 255, 0.7)'
            },
            colorScheme: {
                light: {
                    root: {
                        background: '{zinc.300}',
                        checkedBackground: '{sky.400}',
                        hoverBackground: '{sky.300}',
                        borderColor: '{zinc.400}',
                        color: '{zinc.400}',
                        hoverColor: '{sky.700}',
                        checkedColor: '{zinc.900}',
                        checkedBorderColor: '{sky.500}',
                    },
                    content: {
                        checkedBackground: '{sky.300}'
                    },
                },
            }
        },
/*         tree: {
            colorScheme: {
                light: {
                    root: {
                        background: 'transparent',
                        hoverBackground: '{zinc.400}',
                        color: '{zinc.900}',
                        borderColor: '{zinc.400}',
                        node: {
                            selectedBackground: 'rgba(125, 211, 252, 0.6)',
                            hoverBackground: 'rgba(125, 211, 252, 0.8)',
                        }
                    },
                },
            },
        } */
    },
});
