package main

var instructionsJsonString = []byte(`
[
    {
        "label": "adc",
        "detail": "ADC",
        "documentation": "ADd with Carry",
        "kind": 14
    },
    {
        "label": "sbc",
        "detail": "SBC",
        "documentation": "SuBtract with Carry",
        "kind": 14
    },
    {
        "label": "dec",
        "detail": "DEC",
        "documentation": "DECrease given address",
        "kind": 14
    },
    {
        "label": "dex",
        "detail": "DEX",
        "documentation": "DEcrease register X",
        "kind": 14
    },
    {
        "label": "dey",
        "detail": "DEY",
        "documentation": "DEcrease register Y",
        "kind": 14
    },
    {
        "label": "inc",
        "detail": "INC",
        "documentation": "INCrease gived address",
        "kind": 14
    },
    {
        "label": "inx",
        "detail": "INX",
        "documentation": "INcrease register X",
        "kind": 14
    },
    {
        "label": "iny",
        "detail": "INY",
        "documentation": "INcrease register Y",
        "kind": 14
    },
    {
        "label": "lda",
        "detail": "LDA",
        "documentation": "LoaD register A",
        "kind": 14
    },
    {
        "label": "ldx",
        "detail": "LDX",
        "documentation": "LoaD register X",
        "kind": 14
    },
    {
        "label": "ldy",
        "detail": "LDY",
        "documentation": "LoaD register Y",
        "kind": 14
    },
    {
        "label": "sta",
        "detail": "STA",
        "documentation": "STore register A at address",
        "kind": 14
    },
    {
        "label": "stx",
        "detail": "STX",
        "documentation": "STore register X at address",
        "kind": 14
    },
    {
        "label": "sty",
        "detail": "STY",
        "documentation": "STore register Y at address",
        "kind": 14
    },
    {
        "label": "tax",
        "detail": "TAX",
        "documentation": "Transfer Accumulator to X",
        "kind": 14
    },
    {
        "label": "tay",
        "detail": "TAY",
        "documentation": "Transfer Accumulator to Y",
        "kind": 14
    },
    {
        "label": "txa",
        "detail": "TXA",
        "documentation": "Transfer X to Accumulator",
        "kind": 14
    },
    {
        "label": "tya",
        "detail": "TYA",
        "documentation": "Transfer Y to Accumulator",
        "kind": 14
    },
    {
        "label": "txs",
        "detail": "TXS",
        "documentation": "Transfer X to Stack pointer",
        "kind": 14
    },
    {
        "label": "tsx",
        "detail": "TSX",
        "documentation": "Transfer Stack pointer to X",
        "kind": 14
    },
    {
        "label": "pha",
        "detail": "PHA",
        "documentation": "PusH Accumulator to stack",
        "kind": 14
    },
    {
        "label": "php",
        "detail": "PHP",
        "documentation": "PusH Processor flags to stack",
        "kind": 14
    },
    {
        "label": "pla",
        "detail": "PLA",
        "documentation": "PulL Accumulator from stack",
        "kind": 14
    },
    {
        "label": "plp",
        "detail": "PLP",
        "documentation": "PulL Processor flags from stack",
        "kind": 14
    },
    {
        "label": "cmp",
        "detail": "CMP",
        "documentation": "CoMPare accumulator with address",
        "kind": 14
    },
    {
        "label": "cpx",
        "detail": "CPX",
        "documentation": "ComPare register X with address",
        "kind": 14
    },
    {
        "label": "cpy",
        "detail": "CPY",
        "documentation": "ComPare register Y with address",
        "kind": 14
    },
    {
        "label": "and",
        "detail": "AND",
        "documentation": "Bitwise AND accumulator with address",
        "kind": 14
    },
    {
        "label": "ora",
        "detail": "ORA",
        "documentation": "OR Accumulator with address",
        "kind": 14
    },
    {
        "label": "eor",
        "detail": "EOR",
        "documentation": "Exclusive OR accumulator with address",
        "kind": 14
    },
    {
        "label": "bit",
        "detail": "BIT",
        "documentation": "AND accumulator with address, only changing processor flags",
        "kind": 14
    },
    {
        "label": "sei",
        "detail": "SEI",
        "documentation": "SEt Interrupt disable flag",
        "kind": 14
    },
    {
        "label": "cli",
        "detail": "CLI",
        "documentation": "CLear Interrupt disable flag",
        "kind": 14
    }
]
`)
