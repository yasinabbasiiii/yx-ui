package service

import (
	"encoding/json"
	"errors"
	"sync"

	"x-ui/logger"
	"x-ui/xray"

	"os"
	"strconv"
	"strings"

	"go.uber.org/atomic"
)

var (
	p                 *xray.Process
	lock              sync.Mutex
	isNeedXrayRestart atomic.Bool
	result            string
)

// === Added: runtime client sharing helpers ===
func parseShareEnvMulti() (shareAll bool, fromIDs []int, toSet map[int]bool) {
	shareAll = strings.EqualFold(os.Getenv("XUI_SHARE_ALL"), "true")
	fromIDs = []int{}
	if v := strings.TrimSpace(os.Getenv("XUI_SHARE_FROM_MULTI")); v != "" {
		parts := strings.Split(v, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if n, err := strconv.Atoi(p); err == nil {
				fromIDs = append(fromIDs, n)
			}
		}
	} else {
		fromIDs = []int{1, 21, 78}
	}
	toSet = map[int]bool{}
	if v := strings.TrimSpace(os.Getenv("XUI_SHARE_TO")); v != "" {
		parts := strings.Split(v, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if n, err := strconv.Atoi(p); err == nil {
				toSet[n] = true
			}
		}
	}
	return
}
func shouldShareInbound(shareAll bool, toSet map[int]bool, inboundID int) bool {
	if shareAll {
		return true
	}
	return toSet[inboundID]
}
func appendClientIfNotExistsByEmailOrID(list []interface{}, c map[string]interface{}) []interface{} {
	var email, id string
	if v, ok := c["email"].(string); ok {
		email = v
	}
	if v, ok := c["id"].(string); ok {
		id = v
	}
	for _, it := range list {
		if m, ok := it.(map[string]interface{}); ok {
			if (email != "" && m["email"] == email) || (id != "" && m["id"] == id) {
				return list
			}
		}
	}
	return append(list, c)
}

type XrayService struct {
	inboundService InboundService
	settingService SettingService
	xrayAPI        xray.XrayAPI
}

func (s *XrayService) IsXrayRunning() bool {
	return p != nil && p.IsRunning()
}

func (s *XrayService) GetXrayErr() error {
	if p == nil {
		return nil
	}
	return p.GetErr()
}

func (s *XrayService) GetXrayResult() string {
	if result != "" {
		return result
	}
	if s.IsXrayRunning() {
		return ""
	}
	if p == nil {
		return ""
	}
	result = p.GetResult()
	return result
}

func (s *XrayService) GetXrayVersion() string {
	if p == nil {
		return "Unknown"
	}
	return p.GetVersion()
}

func RemoveIndex(s []interface{}, index int) []interface{} {
	return append(s[:index], s[index+1:]...)
}

func (s *XrayService) GetXrayConfig() (*xray.Config, error) {
	//logger.Debug("GetXrayConfig")

	templateConfig, err := s.settingService.GetXrayConfigTemplate()
	if err != nil {
		logger.Debug("21")
		return nil, err
	}

	xrayConfig := &xray.Config{}
	err = json.Unmarshal([]byte(templateConfig), xrayConfig)
	if err != nil {
		logger.Debug("22")
		return nil, err
	}

	s.inboundService.AddTraffic(nil, nil)

	inbounds, err := s.inboundService.GetAllInbounds()
	if err != nil {
		logger.Debug("23")
		return nil, err
	}
	for _, inbound := range inbounds {
		if !inbound.Enable {
			continue
		}
		// get settings clients
		settings := map[string]interface{}{}
		json.Unmarshal([]byte(inbound.Settings), &settings)
		clients, ok := settings["clients"].([]interface{})
		if ok {
			// check users active or not
			clientStats := inbound.ClientStats
			for _, clientTraffic := range clientStats {
				indexDecrease := 0
				for index, client := range clients {
					c := client.(map[string]interface{})
					if c["email"] == clientTraffic.Email {
						if !clientTraffic.Enable {
							clients = RemoveIndex(clients, index-indexDecrease)
							indexDecrease++
							//logger.Info("Remove Inbound User ", c["email"], " due the expire or traffic limit") //Samyar
						}
					}
				}
			}

			// clear client config for additional parameters
			var final_clients []interface{}
			for _, client := range clients {
				c := client.(map[string]interface{})
				if c["enable"] != nil {
					if enable, ok := c["enable"].(bool); ok && !enable {
						continue
					}
				}
				for key := range c {
					if key != "email" && key != "id" && key != "password" && key != "flow" && key != "method" {
						delete(c, key)
					}
					if c["flow"] == "xtls-rprx-vision-udp443" {
						c["flow"] = "xtls-rprx-vision"
					}
				}
				final_clients = append(final_clients, interface{}(c))
			}

			// === Added: merge clients from source inbound IDs (1,21,78 by default) ===
			shareAll, sourceInboundIDs, targetIDs := parseShareEnvMulti()
			isSource := false
			for _, sid := range sourceInboundIDs {
				if inbound.Id == sid {
					isSource = true
					break
				}
			}
			if !isSource && (shareAll || shouldShareInbound(shareAll, targetIDs, inbound.Id)) {
				for _, sid := range sourceInboundIDs {
					srcInbound, err := s.inboundService.GetInbound(sid)
					if err == nil && srcInbound != nil {
						srcClients, err := s.inboundService.GetClients(srcInbound)
						if err == nil {
							for _, sc := range srcClients {
								cc := map[string]interface{}{}
								if sc.Email != "" {
									cc["email"] = sc.Email
								}
								if sc.ID != "" {
									cc["id"] = sc.ID
								} // ← sc.ID (بزرگ)
								if sc.Password != "" {
									cc["password"] = sc.Password
								}
								if sc.Flow != "" {
									cc["flow"] = sc.Flow
								}
								// توجه: فیلدی به نام Method در model.Client نداریم، پس چیزی ست نمی‌کنیم.

								// فقط کلیدهای مجاز و نرمال‌سازی flow
								cc2 := map[string]interface{}{}
								if v, ok := cc["email"]; ok {
									cc2["email"] = v
								}
								if v, ok := cc["id"]; ok {
									cc2["id"] = v
								}
								if v, ok := cc["password"]; ok {
									cc2["password"] = v
								}
								if v, ok := cc["flow"]; ok {
									if f, ok2 := v.(string); ok2 && f == "xtls-rprx-vision-udp443" {
										cc2["flow"] = "xtls-rprx-vision"
									} else {
										cc2["flow"] = v
									}
								}
								// method نداریم، پس چیزی اضافه نمی‌کنیم

								final_clients = appendClientIfNotExistsByEmailOrID(final_clients, cc2)
							}
						} else {
							logger.Warning("merge clients: failed to get source clients for inbound ", sid, err)
						}
					} else {
						logger.Warning("merge clients: source inbound not found: ", sid, err)
					}
				}
			}
			// === End added ===

			settings["clients"] = final_clients
			modifiedSettings, err := json.MarshalIndent(settings, "", "  ")
			if err != nil {
				logger.Debug("24")
				return nil, err
			}

			inbound.Settings = string(modifiedSettings)
		}

		if len(inbound.StreamSettings) > 0 {
			// Unmarshal stream JSON
			var stream map[string]interface{}
			json.Unmarshal([]byte(inbound.StreamSettings), &stream)

			// Remove the "settings" field under "tlsSettings" and "realitySettings"
			tlsSettings, ok1 := stream["tlsSettings"].(map[string]interface{})
			realitySettings, ok2 := stream["realitySettings"].(map[string]interface{})
			if ok1 || ok2 {
				if ok1 {
					delete(tlsSettings, "settings")
				} else if ok2 {
					delete(realitySettings, "settings")
				}
			}

			delete(stream, "externalProxy")

			newStream, err := json.MarshalIndent(stream, "", "  ")
			if err != nil {
				logger.Debug("24")
				return nil, err
			}
			inbound.StreamSettings = string(newStream)
		}

		inboundConfig := inbound.GenXrayInboundConfig()
		xrayConfig.InboundConfigs = append(xrayConfig.InboundConfigs, *inboundConfig)
	}

	return xrayConfig, nil
}

func (s *XrayService) GetXrayTraffic() ([]*xray.Traffic, []*xray.ClientTraffic, error) {
	if !s.IsXrayRunning() {
		return nil, nil, errors.New("xray is not running")
	}
	s.xrayAPI.Init(p.GetAPIPort())
	defer s.xrayAPI.Close()
	return s.xrayAPI.GetTraffic(true)
}

func (s *XrayService) RestartXray(isForce bool) error {
	lock.Lock()
	defer lock.Unlock()
	logger.Debug("restart xray, force:", isForce)

	xrayConfig, err := s.GetXrayConfig()
	if err != nil {
		return err
	}

	if s.IsXrayRunning() {
		if !isForce && p.GetConfig().Equals(xrayConfig) {
			logger.Debug("It does not need to restart xray")
			return nil
		}
		p.Stop()
	}

	p = xray.NewProcess(xrayConfig)
	result = ""
	err = p.Start()
	if err != nil {
		return err
	}
	return nil
}

func (s *XrayService) StopXray() error {
	lock.Lock()
	defer lock.Unlock()
	logger.Debug("stop xray")
	if s.IsXrayRunning() {
		return p.Stop()
	}
	return errors.New("xray is not running")
}

func (s *XrayService) SetToNeedRestart() {
	logger.Debug("SetToNeedRestart")
	isNeedXrayRestart.Store(true)
}

func (s *XrayService) IsNeedRestartAndSetFalse() bool {
	return isNeedXrayRestart.CompareAndSwap(true, false)
}
