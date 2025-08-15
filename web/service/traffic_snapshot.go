package service

import (
	"x-ui/logger"
)

// SaveTrafficSnapshotOnce:
// یک‌بار مصرف فعلی Xray را می‌گیرد و عیناً مثل Job دوره‌ای، در دیتابیس ذخیره می‌کند.
// این تابع داخل پکیج service است تا import cycle با web/job ایجاد نشود.
func (s *ServerService) SaveTrafficSnapshotOnce() {
	// اگر Xray در حال اجرا نیست، چیزی برای ذخیره نیست
	if !s.xrayService.IsXrayRunning() {
		return
	}

	// reset=true یعنی دلتاهای همین بازه فعلی را بده
	traffics, clientTraffics, err := s.xrayService.GetXrayTraffic()
	if err != nil {
		logger.Warning("get xray traffic failed:", err)
		return
	}

	// ذخیره مصرف اینباندها (منطق جمع/دِدیوپ ایمیلی که قبلاً اصلاح کردیم همین‌جا اعمال می‌شود)
	if err0, needRestart0 := s.inboundService.AddTraffic(traffics, clientTraffics); err0 != nil {
		logger.Warning("add inbound traffic failed:", err0)
	} else if needRestart0 {
		s.xrayService.SetToNeedRestart()
	}

	// // ذخیره مصرف آوت‌باندها
	// var outbound OutboundService
	// if err1, needRestart1 := outbound.AddTraffic(traffics, clientTraffics); err1 != nil {
	// 	logger.Warning("add outbound traffic failed:", err1)
	// } else if needRestart1 {
	// 	s.xrayService.SetToNeedRestart()
	// }
}
