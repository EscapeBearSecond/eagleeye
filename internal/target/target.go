package target

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"github.com/EscapeBearSecond/eagleeye/internal/util"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

// Deprecated: use ProcessV2
//
//	ProcessSync 处理输入的targets（优先作为文件处理）
func ProcessSync(targets []string, excludeTargets ...string) ([]string, error) {
	resultTargets := make([]string, 0, 65536*len(targets))
	for _, target := range targets {
		expands, err := Expand(target)
		if err == nil {
			resultTargets = append(resultTargets, expands...)
			continue
		}

		info, err := os.Stat(target)
		if err != nil {
			return nil, fmt.Errorf("get target file stat failed: %w", err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("unsupported target file type: %w", err)
		}
		tf, err := os.Open(target)
		if err != nil {
			return nil, fmt.Errorf("open target file failed: %w", err)
		}
		scanner := bufio.NewScanner(tf)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			expands, err := Expand(scanner.Text())
			if err != nil {
				return nil, fmt.Errorf("expand [%s] from target file failed: %w", scanner.Text(), err)
			}

			resultTargets = append(resultTargets, expands...)
		}
	}

	if len(excludeTargets) != 0 {
		excluds, err := ProcessSync(excludeTargets)
		if err != nil {
			return nil, fmt.Errorf("process exclude targets failed: %w", err)
		}

		excludeMap := lo.SliceToMap(excluds, func(exlude string) (string, struct{}) {
			return exlude, struct{}{}
		})

		resultTargets = lo.Filter(resultTargets, func(target string, _ int) bool {
			_, contained := excludeMap[target]
			return !contained
		})
	}

	return lo.Uniq(resultTargets), nil
}

// ProcessAsync 处理输入的targets（优化效率）
func ProcessAsync(targets []string, excludeTargets ...string) ([]string, error) {
	resultTargets := make([]string, 0, 65536*len(targets))

	m := sync.Mutex{}
	eg, ctx := errgroup.WithContext(context.Background())
	for _, target := range targets {
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			expands, err := Expand(target)
			if err == nil {
				select {
				case <-ctx.Done():
					return nil
				default:
				}
				m.Lock()
				resultTargets = append(resultTargets, expands...)
				m.Unlock()
				return nil
			}
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			info, err := os.Stat(target)
			if err != nil {
				return fmt.Errorf("get target file stat failed: %w", err)
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("unsupported target file type: %w", err)
			}
			tf, err := os.Open(target)
			if err != nil {
				return fmt.Errorf("open target file failed: %w", err)
			}
			scanner := bufio.NewScanner(tf)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return nil
				default:
				}
				expands, err := Expand(scanner.Text())
				if err != nil {
					return fmt.Errorf("expand [%s] from target file failed: %w", scanner.Text(), err)
				}

				select {
				case <-ctx.Done():
					return nil
				default:
				}
				m.Lock()
				resultTargets = append(resultTargets, expands...)
				m.Unlock()
			}
			return nil
		})
	}

	var excludeMap map[string]struct{}
	if len(excludeTargets) != 0 {
		eg.Go(func() error {
			excluds, err := ProcessAsync(excludeTargets)
			if err != nil {
				return fmt.Errorf("process exclude targets failed: %w", err)
			}

			excludeMap = lo.SliceToMap(excluds, func(exlude string) (string, struct{}) {
				return exlude, struct{}{}
			})
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if excludeMap != nil {
		resultTargets = lo.Filter(resultTargets, func(target string, _ int) bool {
			_, contained := excludeMap[target]
			return !contained
		})
	}

	return lo.Uniq(resultTargets), nil
}

// Expand 扩展目标
//
// CIDR: 192.168.1.0/24
//
// IPRange: 192.168.1.1-192.168.2.3
//
// IP: 192.168.1.123
//
// HostPort: 192.168.1.156:8090
func Expand(target string) ([]string, error) {
	switch {
	case util.IsCIDR(target):
		return util.ExpandCIDR(target), nil
	case util.IsIPRange(target):
		return util.ExpandIPRange(target), nil
	case util.IsIP(target):
		return []string{target}, nil
	case util.IsHostPort(target):
		return []string{target}, nil
	default:
		return nil, fmt.Errorf("invalid target: %s", target)
	}
}

func ShouldSkip(target string, ports ...string) bool {
	// 如果ports不为空，并且target为ip:port格式
	if util.IsHostPort(target) && len(ports) != 0 {
		// 获取target中的port值
		_, port, _ := net.SplitHostPort(target)

		// 如果扫描的port不在指定的ports中，则跳过
		if !lo.Contains(ports, port) {
			return true
		}
	}
	return false
}

func SplitBySize(targets []string, size uint32) ([][]string, error) {
	if len(targets) == 0 {
		return nil, errors.New("empty targets")
	}

	if size == 0 {
		return nil, errors.New("invalid size")
	}

	var sizeSum uint32
	for _, target := range targets {
		switch {
		case util.IsCIDR(target):
			_, size := util.CIDRSize(target)
			sizeSum += size
		case util.IsIPRange(target):
			_, size := util.IPRangeSize(target)
			sizeSum += size
		case util.IsIP(target):
			sizeSum++
		case util.IsHostPort(target):
			sizeSum++
		default:
		}
	}

	if sizeSum == 0 {
		return nil, errors.New("empty targets")
	}

	if sizeSum <= size {
		return [][]string{targets}, nil
	}

	n := sizeSum / size
	remainder := sizeSum % size
	if remainder != 0 {
		n++
	}

	return SplitN(targets, int(n))
}

func SplitN(targets []string, n int) ([][]string, error) {
	if n <= 0 {
		return nil, errors.New("n must be greater than 0")
	}

	if n == 1 {
		return [][]string{targets}, nil
	}

	chs := make([]chan string, 0, n)
	for range n {
		chs = append(chs, make(chan string, 5))
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	go func() {
		defer func() {
			for _, ch := range chs {
				close(ch)
			}
		}()

		for _, target := range targets {
			switch {
			case util.IsCIDR(target):
				ip, size := util.CIDRSize(target)
				perSize := size / uint32(n)
				remainder := size % uint32(n)
				for i := uint32(0); i < size-remainder; i += perSize {
					if i == size-remainder-perSize {
						chs[i/perSize] <- fmt.Sprintf("%s-%s",
							util.OffsetIP(ip, i), util.OffsetIP(ip, i+perSize+remainder-1))
						continue
					}
					chs[i/perSize] <- fmt.Sprintf("%s-%s",
						util.OffsetIP(ip, i), util.OffsetIP(ip, i+perSize-1))
				}
			case util.IsIPRange(target):
				ip, size := util.IPRangeSize(target)
				perSize := size / uint32(n)
				remainder := size % uint32(n)
				for i := uint32(0); i < size-remainder; i += perSize {
					if i == size-remainder-perSize {
						chs[i/perSize] <- fmt.Sprintf("%s-%s",
							util.OffsetIP(ip, i), util.OffsetIP(ip, i+perSize+remainder-1))
						continue
					}
					chs[i/perSize] <- fmt.Sprintf("%s-%s",
						util.OffsetIP(ip, i), util.OffsetIP(ip, i+perSize-1))
				}
			case util.IsIP(target):
				chs[r.Intn(n)] <- target
			case util.IsHostPort(target):
				chs[r.Intn(n)] <- target
			default:
			}
		}
	}()

	results := make([][]string, 0, n)
	for range n {
		results = append(results, make([]string, 0))
	}

	wg := sync.WaitGroup{}
	for i, ch := range chs {
		wg.Add(1)
		go func(ch chan string) {
			defer wg.Done()
			for target := range ch {
				results[i] = append(results[i], target)
			}
		}(ch)
	}

	wg.Wait()

	r.Shuffle(len(results), func(i, j int) {
		results[i], results[j] = results[j], results[i]
	})

	return results, nil
}
