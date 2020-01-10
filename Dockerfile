FROM alpine
LABEL maintainers="Xnile"
LABEL description="KVM CSI Driver"

RUN apk add util-linux e2fsprogs
COPY kvm-csi-driver /kvm-csi-driver
ENTRYPOINT ["/kvm-csi-driver"]