package core

import "fmt"

type Ffx2Duplicate struct {
	duplicates *Duplicate
}

var ffx2DuplicateInstance *Ffx2Duplicate

func NewFfx2Duplicate() *Ffx2Duplicate {
	if ffx2DuplicateInstance == nil {
		fmt.Println("NewFfx2Duplicate")

		duplicates := NewDuplicate()

		duplicates.AddFromSpaceSeparatedString("klyt0500", "klyt1600")
		duplicates.AddFromSpaceSeparatedString("bika0700", "bika1100")
		duplicates.AddFromSpaceSeparatedString("bsil0300", "bsil0500")
		duplicates.AddFromSpaceSeparatedString("hiku3000", "hiku3500")
		duplicates.AddFromSpaceSeparatedString("klyt0900", "klyt1700")
		duplicates.AddFromSpaceSeparatedString("hiku2903", "hiku3400")
		duplicates.AddFromSpaceSeparatedString("dnfr8000", "dnfr8001")
		duplicates.AddFromSpaceSeparatedString("stbv1200", "stbv1300")
		duplicates.AddFromSpaceSeparatedString("hiku2800", "hiku3200")

		duplicates.AddFromSpaceSeparatedString("dnfr0100",
			"dnfr0300 dnfr0500 dnfr0700 dnfr0900 dnfr1100 dnfr1300 dnfr1500 dnfr1700 dnfr1800 dnfr2000 dnfr2100 dnfr2300 dnfr2500 dnfr2700 dnfr2900 dnfr3100 dnfr3300 dnfr3500 dnfr3700 dnfr4000 dnfr4100 dnfr4300 dnfr4500 dnfr4700 dnfr4900 dnfr5100 dnfr5300 dnfr5500 dnfr5700 dnfr5800 dnfr6000 dnfr6100 dnfr6300 dnfr6500 dnfr6700 dnfr6800 dnfr6900 dnfr7000 dnfr7100 dnfr7200 dnfr7300 dnfr7500 dnfr7600 dnfr7700 dnfr7800 dnfr7900 lmys0000 spdn0100 spdn0400 spdn0401 spdn0402 spdn0403 spdn0500 spdn0501 spdn0502 spdn0503 spdn0600 spdn0601 spdn0602 spdn0603 spdn0700")

		duplicates.AddFromSpaceSeparatedString("bika07_235",
			"bika07_236 bika07_237 bika07_238 bika07_240 bika07_241 bika07_242 bika07_243 bika07_250 bika07_251 bika07_252 bika07_253 bika08_230 bika08_231 bika08_232 bika08_233 bika08_235 bika08_236 bika08_237 bika08_238 bika08_240 bika08_241 bika08_242 bika08_243 bika08_250 bika08_251 bika08_252 bika08_253 bika09_235 bika09_236 bika09_237 bika09_238 bika09_240 bika09_241 bika09_242 bika09_243 bika09_250 bika09_251 bika09_252 bika09_253 bika13_240 bika13_241 bika13_242 bika13_243 bika13_250 bika13_251 bika13_252 bika13_253 bika14_235 bika14_236 bika14_237 bika14_238 bika14_240 bika14_241 bika14_242 bika14_243 bika14_250 bika14_251 bika14_252 bika14_253 bsil05_230 bsil05_231 bsil05_232 bsil05_233 bsil05_235 bsil05_236 bsil05_237 bsil05_238 bsil05_240 bsil05_241 bsil05_242 bsil05_243 bsil05_250 bsil05_251 bsil05_252 bsil05_253 bsil07_230 bsil07_231 bsil07_232 bsil07_233 bsil07_235 bsil07_236 bsil07_237 bsil07_238 bsil07_240 bsil07_241 bsil07_242 bsil07_243 bsil07_250 bsil07_251 bsil07_252 bsil07_253 genk00_230 genk00_231 genk00_232 genk00_233 genk00_235 genk00_236 genk00_237 genk00_238 genk00_240 genk00_241 genk00_242 genk00_243 genk00_245 genk00_246 genk00_247 genk00_248 genk00_250 genk00_251 genk00_252 genk00_253 genk15_245 genk15_246 genk15_247 genk15_248 genk16_230 genk16_231 genk16_232 genk16_233 genk16_235 genk16_236 genk16_237 genk16_238 genk16_240 genk16_241 genk16_242 genk16_243 genk16_245 genk16_246 genk16_247 genk16_248 genk16_250 genk16_251 genk16_252 genk16_253 kami00_230 kami00_231 kami00_232 kami00_233 kami00_235 kami00_236 kami00_237 kami00_238 kami00_240 kami00_241 kami00_242 kami00_243 kami00_245 kami00_246 kami00_247 kami00_248 kami00_250 kami00_251 kami00_252 kami00_253 kami03_230 kami03_231 kami03_232 kami03_233 kami03_235 kami03_236 kami03_237 kami03_238 kami03_240 kami03_241 kami03_242 kami03_243 kami03_245 kami03_246 kami03_247 kami03_248 kami03_250 kami03_251 kami03_252 kami03_253 kino04_100 kino04_101 kino04_102 kino04_105 kino04_106 kino04_107 kino04_110 kino04_111 kino04_112 kino04_120 kino04_121 kino04_122 kino04_230 kino04_231 kino04_232 kino04_233 kino04_235 kino04_236 kino04_237 kino04_238 kino04_240 kino04_241 kino04_242 kino04_243 kino04_250 kino04_251 kino04_252 kino04_253 klyt00_230 klyt00_231 klyt00_232 klyt00_233 klyt00_240 klyt00_241 klyt00_242 klyt00_243 klyt00_250 klyt00_251 klyt00_252 klyt00_253 mcfr00_230 mcfr00_231 mcfr00_232 mcfr00_233 mcfr00_235 mcfr00_236 mcfr00_237 mcfr00_238 mcfr00_240 mcfr00_241 mcfr00_242 mcfr00_243 mcfr00_250 mcfr00_251 mcfr00_252 mcfr00_253 mcfr01_230 mcfr01_231 mcfr01_232 mcfr01_233 mcfr01_235 mcfr01_236 mcfr01_237 mcfr01_238 mcfr01_240 mcfr01_241 mcfr01_242 mcfr01_243 mcfr01_250 mcfr01_251 mcfr01_252 mcfr01_253 mcfr02_230 mcfr02_231 mcfr02_232 mcfr02_233 mcfr02_235 mcfr02_236 mcfr02_237 mcfr02_238 mcfr02_240 mcfr02_241 mcfr02_242 mcfr02_243 mcfr02_250 mcfr02_251 mcfr02_252 mcfr02_253 nagi00_100 nagi00_101 nagi00_102 nagi00_105 nagi00_106 nagi00_107 nagi00_110 nagi00_111 nagi00_112 nagi00_120 nagi00_121 nagi00_122 nagi00_130 nagi00_131 nagi00_132 nagi00_133 nagi00_135 nagi00_136 nagi00_137 nagi00_138 nagi00_140 nagi00_141 nagi00_142 nagi00_143 nagi00_150 nagi00_151 nagi00_152 nagi00_153 nagi00_230 nagi00_231 nagi00_232 nagi00_235 nagi00_236 nagi00_237 nagi00_240 nagi00_241 nagi00_242 nagi00_250 nagi00_251 nagi00_252",
		)
		ffx2DuplicateInstance = &Ffx2Duplicate{
			duplicates: duplicates,
		}
	}

	return ffx2DuplicateInstance
}

func (f *Ffx2Duplicate) TryFind(key string) []string {
	return f.duplicates.Find(key)
}
