// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/cloud-sql-proxy-operator/internal/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type upgradeDefaultProxyOnStartup struct {
	c client.Client
}

func (c *upgradeDefaultProxyOnStartup) findProxyNeedingUpdate(ctx context.Context) error {

	l := &v1alpha1.AuthProxyWorkloadList{}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := c.c.List(ctx, l, client.Continue(l.Continue))
			if err != nil {
				return fmt.Errorf("can't list AuthProxyWorkload on startup, %v", err)
			}

			for _, p := range l.Items {
				useDefaultImage := p.Spec.AuthProxyContainer == nil || p.Spec.AuthProxyContainer.Image == ""

				if !useDefaultImage {
					continue
				}

				// If an APW has a default image, then perform an "update" on it so that
				// the reconcile function runs and triggers the appropriate rolling updates.
				log.FromContext(ctx).Info(fmt.Sprintf("Upgrading workload default images for %s/%s", p.Namespace, p.Namespace))
				err = c.c.Update(ctx, &p)
			}
			if l.Continue == "" {
				return nil
			}
		}
	}

}

func (c *upgradeDefaultProxyOnStartup) Start(ctx context.Context) error {
	go c.findProxyNeedingUpdate(ctx)
	return nil
}

func (c *upgradeDefaultProxyOnStartup) NeedLeaderElection() bool {
	return true
}
