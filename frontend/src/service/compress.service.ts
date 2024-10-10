import { Injectable } from '@angular/core';
import { Compress } from '../../wailsjs/go/services/CompressService';
import { lib } from '../../wailsjs/go/models';

@Injectable({
  providedIn: 'root'
})
export class CompressService {
  async compress(fileInfo: lib.FileInfo) {
    Compress(fileInfo);
  }
}
